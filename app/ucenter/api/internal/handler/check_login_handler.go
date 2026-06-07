package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CheckLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("x-auth-token")
		resp, err := logic.NewCheckLoginLogic(r.Context(), svcCtx).CheckLogin(token)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
