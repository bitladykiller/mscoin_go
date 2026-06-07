package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SendWithdrawCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := logic.NewSendWithdrawCodeLogic(r.Context(), svcCtx).SendCode()
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
