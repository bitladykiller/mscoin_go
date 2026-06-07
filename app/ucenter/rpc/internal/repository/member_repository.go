// Package repository 定义会员仓储层。
//
// MemberRepository 封装会员表（member）的所有数据库操作。
// 仓储层职责：
//   - 提供数据访问接口，隔离业务逻辑与数据库细节
//   - 实现查询、插入、更新等基本 CRUD 操作
//   - 不包含复杂业务逻辑，业务逻辑由 domain/service 层处理
//
// 设计原则：
//   - 单一职责：每个仓储只负责一张主表
//   - 接口隔离：仓储只暴露必要的数据访问方法
//   - 错误处理：区分"未找到"和"查询失败"
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// MemberRepository 会员仓储
// 负责会员数据的持久化操作
//
// 使用 sqlx 进行数据库操作，支持：
//   - 结构体映射：自动将查询结果映射到结构体
//   - 参数绑定：安全的参数化查询，防止 SQL 注入
//   - 事务支持：可与其他仓储共享事务
type MemberRepository struct {
	db *sqlx.DB // 数据库连接池
}

// NewMemberRepository 创建会员仓储实例
// 参数 db 为数据库连接池，由 ServiceContext 提供
func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

// FindByPhone 根据手机号查询会员
// 用于登录验证和注册重复检查
//
// 返回值：
//   - 找到会员：返回会员对象，nil 错误
//   - 未找到：返回 nil，nil 错误（非错误场景）
//   - 查询失败：返回 nil，错误信息
func (r *MemberRepository) FindByPhone(ctx context.Context, phone string) (*model.Member, error) {
	var member model.Member
	err := r.db.GetContext(ctx, &member, "SELECT * FROM member WHERE mobile_phone=? LIMIT 1", phone)
	if errors.Is(err, sql.ErrNoRows) {
		// 未找到会员，返回 nil 而非错误
		// 这是业务逻辑需要区分的场景
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query member by phone: %w", err)
	}
	return &member, nil
}

// FindByID 根据会员 ID 查询会员
// 用于会员信息查询和身份验证
//
// 返回值：
//   - 找到会员：返回会员对象，nil 错误
//   - 未找到：返回 nil，nil 错误
//   - 查询失败：返回 nil，错误信息
func (r *MemberRepository) FindByID(ctx context.Context, memberID int64) (*model.Member, error) {
	var member model.Member
	err := r.db.GetContext(ctx, &member, "SELECT * FROM member WHERE id=? LIMIT 1", memberID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query member by id: %w", err)
	}
	return &member, nil
}

// UpdateLoginCount 更新会员登录次数
// 每次登录成功后调用，用于统计会员活跃度
//
// 参数：
//   - ctx: 请求上下文
//   - id: 会员 ID
//   - step: 增加步数，通常为 1
//
// 注意：此操作通常异步执行，不阻塞登录响应
func (r *MemberRepository) UpdateLoginCount(ctx context.Context, id int64, step int) error {
	if _, err := r.db.ExecContext(ctx, "UPDATE member SET login_count = login_count + ? WHERE id = ?", step, id); err != nil {
		return fmt.Errorf("update member login count: %w", err)
	}
	return nil
}

// Save 保存会员记录
// 用于注册新会员
//
// 参数 member 必须包含所有必填字段：
//   - 手机号、用户名、密码、盐值等
//   - 注册时间由调用方设置
//
// 注意：由于旧版表结构较宽，需要填写所有字段
// 使用默认值填充非必填字段，保持数据完整性
func (r *MemberRepository) Save(ctx context.Context, member *model.Member) error {
	if member == nil {
		return errors.New("member is nil")
	}

	// 会员表字段较多，使用完整的 INSERT 语句
	// 按表结构顺序填写所有字段，避免遗漏
	const query = `
INSERT INTO member (
	ali_no, qr_code_url, appeal_success_times, appeal_times, application_time, avatar, bank, branch, card_no,
	certified_business_apply_time, certified_business_check_time, certified_business_status, channel_id, email,
	first_level, google_date, google_key, google_state, id_number, inviter_id, is_channel, jy_password,
	last_login_time, city, country, district, province, login_count, login_lock, margin, member_level,
	mobile_phone, password, promotion_code, publish_advertise, real_name, real_name_status, registration_time,
	salt, second_level, sign_in_ability, status, third_level, token, token_expire_time, transaction_status,
	transaction_time, transactions, username, qr_we_code_url, wechat, local, integration, member_grade_id,
	kyc_status, generalize_total, inviter_parent_id, super_partner, kick_fee, power, team_level, team_power,
	member_level_id
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?,
	?
)`
	// 保存会员的所有字段到数据库
	_, err := r.db.ExecContext(
		ctx,
		query,
		member.AliNo, member.QrCodeUrl, member.AppealSuccessTimes, member.AppealTimes, member.ApplicationTime, member.Avatar, member.Bank, member.Branch, member.CardNo,
		member.CertifiedBusinessApplyTime, member.CertifiedBusinessCheckTime, member.CertifiedBusinessStatus, member.ChannelId, member.Email,
		member.FirstLevel, member.GoogleDate, member.GoogleKey, member.GoogleState, member.IdNumber, member.InviterId, member.IsChannel, member.JyPassword,
		member.LastLoginTime, member.City, member.Country, member.District, member.Province, member.LoginCount, member.LoginLock, member.Margin, member.MemberLevel,
		member.MobilePhone, member.Password, member.PromotionCode, member.PublishAdvertise, member.RealName, member.RealNameStatus, member.RegistrationTime,
		member.Salt, member.SecondLevel, member.SignInAbility, member.Status, member.ThirdLevel, member.Token, member.TokenExpireTime, member.TransactionStatus,
		member.TransactionTime, member.Transactions, member.Username, member.QrWeCodeUrl, member.Wechat, member.Local, member.Integration, member.MemberGradeId,
		member.KycStatus, member.GeneralizeTotal, member.InviterParentId, member.SuperPartner, member.KickFee, member.Power, member.TeamLevel, member.TeamPower,
		member.MemberLevelId,
	)
	if err != nil {
		return fmt.Errorf("save member: %w", err)
	}
	return nil
}