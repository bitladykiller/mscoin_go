// Package service 定义提现领域服务。
//
// WithdrawService 是提现管理的核心领域服务，负责：
//   - 提现申请：验证验证码、交易密码，冻结余额，创建记录
//   - 验证码发送：生成并缓存提现验证码
//   - 提现地址查询：查询会员保存的提现地址
//   - 提现记录查询：查询会员的提现历史
//
// 提现流程（Apply 方法）：
//  1. 验证请求参数（金额、地址、验证码等）
//  2. 查询会员信息，获取手机号
//  3. 验证 Redis 中的验证码
//  4. 验证交易密码
//  5. 在事务中执行：
//     - 使用 FOR UPDATE 锁定钱包行
//     - 检查余额是否充足
//     - 冻结余额（Balance -> FrozenBalance）
//     - 创建提现记录
//  6. 发布 Kafka 事件通知下游处理
//
// 事务安全设计：
//   - 使用 FOR UPDATE 行锁防止并发提现超扣
//   - 余额冻结在 SQL 中原子执行
//   - 提现记录创建和余额冻结在同一事务中
//
// Kafka 事件发布：
//   - 事件类型：提现申请
//   - 事件内容：提现记录 JSON
//   - 消费者：jobcenter 执行实际链上转账
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/model"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/mq/kafka"
)

// withdrawCacheKey 提现验证码缓存键前缀
// 格式：WITHDRAW::{phone}
// 用于缓存发送给指定手机号的提现验证码
const withdrawCacheKey = "WITHDRAW::"

// --- 仓储接口定义 ---

// withdrawMemberRepository 提现服务所需的会员仓储接口
// 只定义提现服务需要的方法，遵循接口隔离原则
type withdrawMemberRepository interface {
	FindByID(ctx context.Context, memberID int64) (*model.Member, error)
}

// withdrawWalletRepository 提现服务所需的钱包仓储接口
// 包含普通查询和事务查询方法
type withdrawWalletRepository interface {
	FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error)
	FindByMemberIDAndCoinNameForUpdate(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error)
	FreezeBalance(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string, amount float64) error
}

// withdrawAddressRepository 提现服务所需的地址仓储接口
// 用于查询会员保存的提现地址
type withdrawAddressRepository interface {
	FindByMemberIDAndCoinID(ctx context.Context, memberID int64, coinID int64) ([]*model.MemberAddress, error)
}

// withdrawRecordRepository 提现服务所需的记录仓储接口
// 用于保存和查询提现记录
type withdrawRecordRepository interface {
	FindByMemberID(ctx context.Context, memberID int64, page int64, pageSize int64) ([]*model.WithdrawRecord, int64, error)
	Save(ctx context.Context, exec mysqlx.ExtContext, record *model.WithdrawRecord) error
}

// withdrawCache 提现服务所需的缓存接口
// 用于存储验证码
type withdrawCache interface {
	GetCtx(ctx context.Context, key string, value any) error
	SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error
}

// WithdrawService 聚合了迁移后的提现读侧工作流。
//
// 读侧流程为何集中在此：
//   - gRPC 处理器和未来的异步工作器需要相同的映射规则
//   - 将市场币种丰富逻辑保持在仓储代码之外，以保持持久层与编排层之间的清晰分层
//   - 基于 Redis 的验证码处理属于领域服务层，而非传输适配器
//   - 写侧提现申请逻辑也必须保留在此，以便事务编排、缓存验证和 Kafka 派发保持可复用
type WithdrawService struct {
	memberRepo  withdrawMemberRepository   // 会员仓储
	walletRepo  withdrawWalletRepository   // 钱包仓储
	addressRepo withdrawAddressRepository  // 地址仓储
	recordRepo  withdrawRecordRepository   // 记录仓储
	cache       withdrawCache              // 缓存客户端
	txManager   mysqlx.TxManager           // 事务管理器
	queue       kafka.Producer             // Kafka 生产者
}

// NewWithdrawService 创建提现服务实例
// 参数通过依赖注入，便于单元测试时 Mock
func NewWithdrawService(
	memberRepo withdrawMemberRepository,
	walletRepo withdrawWalletRepository,
	addressRepo withdrawAddressRepository,
	recordRepo withdrawRecordRepository,
	cache withdrawCache,
	txManager mysqlx.TxManager,
	queue kafka.Producer,
) *WithdrawService {
	return &WithdrawService{
		memberRepo:  memberRepo,
		walletRepo:  walletRepo,
		addressRepo: addressRepo,
		recordRepo:  recordRepo,
		cache:       cache,
		txManager:   txManager,
		queue:       queue,
	}
}

// FindAddressByCoinID 根据币种 ID 查询会员提现地址
// 用于提现页面展示会员保存的提现地址列表
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//   - coinID: 币种 ID
//
// 返回：该会员该币种下的所有提现地址
func (s *WithdrawService) FindAddressByCoinID(ctx context.Context, memberID int64, coinID int64) ([]*withdrawpb.AddressSimple, error) {
	list, err := s.addressRepo.FindByMemberIDAndCoinID(ctx, memberID, coinID)
	if err != nil {
		return nil, err
	}

	// 转换为 protobuf 响应
	resp := make([]*withdrawpb.AddressSimple, 0, len(list))
	for _, item := range list {
		resp = append(resp, item.ToProto())
	}
	return resp, nil
}

// SendCode 发送提现验证码
// 生成验证码并缓存到 Redis
//
// 验证码规则：
//   - 长度：6 位数字
//   - 有效期：5 分钟
//   - 缓存键：WITHDRAW::{phone}
//
// 注意：实际发送短信由外部服务完成，本方法只负责生成和缓存
//
// 参数：
//   - ctx: 请求上下文
//   - phone: 手机号
//
// 返回：错误信息
func (s *WithdrawService) SendCode(ctx context.Context, phone string) error {
	if phone == "" {
		return errors.New("phone is required")
	}

	// 生成 6 位数字验证码
	// 使用加密安全的随机数生成器
	code, err := generateNumericCode(6)
	if err != nil {
		return errors.New("generate verification code failed")
	}

	// 缓存验证码，有效期 5 分钟
	// 后续提现申请时需要验证此验证码
	if err := s.cache.SetWithExpireCtx(ctx, withdrawCacheKey+phone, code, 5*time.Minute); err != nil {
		return errors.New("send withdraw verification code failed")
	}
	return nil
}

// Apply 执行写侧提现申请工作流。
//
// 当前迁移策略：
//   - 首先验证 Redis 验证码和交易密码
//   - 在一个 SQL 事务内锁定会员钱包行
//   - 冻结请求的余额并持久化一条提现记录
//   - 在提交前发布 Kafka 事件，以便在此迁移阶段消息投递失败仍能回滚余额冻结
//
// 这在完整的 outbox/consumer 重构仍在 `jobcenter` 中待完成时，
// 保持了与旧版原子意图的一致性。
//
// 事务处理流程：
//  1. 开启事务
//  2. 使用 FOR UPDATE 锁定钱包行
//  3. 检查余额是否充足
//  4. 冻结余额（Balance -> FrozenBalance）
//  5. 创建提现记录
//  6. 发布 Kafka 事件
//  7. 提交事务
//
// 参数：
//   - ctx: 请求上下文
//   - req: 提现请求，包含用户 ID、币种、金额、地址、验证码、交易密码等
//
// 返回：错误信息
func (s *WithdrawService) Apply(ctx context.Context, req *withdrawpb.WithdrawReq) error {
	// 验证请求参数
	if req == nil {
		return errors.New("withdraw request is required")
	}
	if err := validateWithdrawApplyRequest(req); err != nil {
		return err
	}

	// 查询会员信息
	// 获取手机号用于验证验证码
	member, err := s.memberRepo.FindByID(ctx, req.UserId)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.New("member not found")
	}
	if strings.TrimSpace(member.MobilePhone) == "" {
		return errors.New("member phone is unavailable")
	}

	// 验证 Redis 中的验证码
	// 验证码有效期 5 分钟
	var cachedCode string
	if err := s.cache.GetCtx(ctx, withdrawCacheKey+member.MobilePhone, &cachedCode); err != nil {
		return errors.New("verification code unavailable")
	}
	if cachedCode != req.Code {
		return errors.New("verification code mismatch")
	}

	// 验证交易密码
	// 交易密码是提现的第二重验证
	if member.JyPassword != req.JyPassword {
		return errors.New("wrong transaction password")
	}

	// 在事务中执行提现申请
	// 使用事务管理器确保原子性
	return s.txManager.WithinTx(ctx, func(exec mysqlx.ExtContext) error {
		// 使用 FOR UPDATE 锁定钱包行
		// 防止并发提现请求同时冻结相同的余额快照
		wallet, err := s.walletRepo.FindByMemberIDAndCoinNameForUpdate(ctx, exec, req.UserId, req.Unit)
		if err != nil {
			return err
		}
		if wallet == nil {
			return errors.New("wallet not found")
		}

		// 检查余额是否充足
		// 可用余额必须大于等于提现金额
		if wallet.Balance < req.Amount {
			return errors.New("insufficient balance")
		}

		// 冻结余额
		// 将提现金额从可用余额转移到冻结余额
		// 使用 SQL 原子操作，避免竞态条件
		if err := s.walletRepo.FreezeBalance(ctx, exec, req.UserId, req.Unit, req.Amount); err != nil {
			return err
		}

		// 创建提现记录
		// 状态为 Processing，等待下游处理
		record := model.NewWithdrawRecordForApply(time.Now(), wallet, req)
		if err := s.recordRepo.Save(ctx, exec, record); err != nil {
			return err
		}

		// 发布 Kafka 事件
		// 通知 jobcenter 执行实际链上转账
		// 使用用户 ID 作为分区键，保证同一用户的提现顺序
		message, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("marshal withdraw event: %w", err)
		}
		if err := s.queue.PushWithKey(ctx, strconv.FormatInt(req.UserId, 10), string(message)); err != nil {
			return fmt.Errorf("publish withdraw event: %w", err)
		}
		return nil
	})
}

// FindRecordList 查询会员提现记录列表
// 用于提现历史页面展示
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//   - page: 页码，从 1 开始
//   - pageSize: 每页条数
//   - findCoin: 获取币种信息的函数（从 Market RPC 获取）
//
// 返回：
//   - list: 提现记录列表
//   - total: 总记录数（用于分页计算）
//   - error: 错误信息
func (s *WithdrawService) FindRecordList(ctx context.Context, memberID int64, page int64, pageSize int64, findCoin func(context.Context, int64) (*marketpb.Coin, error)) ([]*withdrawpb.WithdrawRecord, int64, error) {
	// 从仓储查询提现记录
	// 支持分页
	list, total, err := s.recordRepo.FindByMemberID(ctx, memberID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 转换为 protobuf 响应
	// 每条记录需要获取对应的币种信息
	resp := make([]*withdrawpb.WithdrawRecord, 0, len(list))
	for _, record := range list {
		coin, err := findCoin(ctx, record.CoinId)
		if err != nil {
			return nil, 0, err
		}
		if coin == nil {
			return nil, 0, fmt.Errorf("coin %d not found", record.CoinId)
		}
		resp = append(resp, record.ToProto(coin))
	}

	return resp, total, nil
}

// validateWithdrawApplyRequest 验证提现申请请求参数
// 确保所有必填字段都已正确填写
//
// 验证规则：
//   - 用户 ID：必须大于 0
//   - 币种：不能为空
//   - 地址：不能为空
//   - 金额：必须大于 0
//   - 手续费：不能为负数，不能超过金额
//   - 交易密码：不能为空
//   - 验证码：不能为空
func validateWithdrawApplyRequest(req *withdrawpb.WithdrawReq) error {
	if req.UserId <= 0 {
		return errors.New("user id is required")
	}
	if strings.TrimSpace(req.Unit) == "" {
		return errors.New("coin unit is required")
	}
	if strings.TrimSpace(req.Address) == "" {
		return errors.New("withdraw address is required")
	}
	if req.Amount <= 0 {
		return errors.New("withdraw amount must be greater than zero")
	}
	if req.Fee < 0 {
		return errors.New("withdraw fee cannot be negative")
	}
	if req.Fee > req.Amount {
		return errors.New("withdraw fee exceeds amount")
	}
	if strings.TrimSpace(req.JyPassword) == "" {
		return errors.New("transaction password is required")
	}
	if strings.TrimSpace(req.Code) == "" {
		return errors.New("verification code is required")
	}
	return nil
}
