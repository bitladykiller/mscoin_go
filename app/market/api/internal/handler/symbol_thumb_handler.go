package handler

import (
	"net/http"

	"mscoin_go/app/market/api/internal/logic"
	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	"mscoin_go/pkg/httputil"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SymbolThumbHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &types.MarketReq{IP: httputil.ClientIP(r)}
		resp, err := logic.NewSymbolThumbLogic(r.Context(), svcCtx).SymbolThumb(req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
