package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

type MemberRepository struct {
	db *sqlx.DB
}

func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) FindByPhone(ctx context.Context, phone string) (*model.Member, error) {
	var member model.Member
	err := r.db.GetContext(ctx, &member, "SELECT * FROM member WHERE mobile_phone=? LIMIT 1", phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query member by phone: %w", err)
	}
	return &member, nil
}

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

func (r *MemberRepository) UpdateLoginCount(ctx context.Context, id int64, step int) error {
	if _, err := r.db.ExecContext(ctx, "UPDATE member SET login_count = login_count + ? WHERE id = ?", step, id); err != nil {
		return fmt.Errorf("update member login count: %w", err)
	}
	return nil
}

func (r *MemberRepository) Save(ctx context.Context, member *model.Member) error {
	if member == nil {
		return errors.New("member is nil")
	}

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
	?, ?, ?, ?, ?, ?, ?, ?,
	?, ?, ?, ?, ?, ?, ?, ?,
	?
)`

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
