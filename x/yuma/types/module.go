package types

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

const (
	// ModuleName 模块名称
	ModuleName = "yuma"

	// RouterKey 路由键
	RouterKey = ModuleName

	// QuerierRoute 查询路由
	QuerierRoute = ModuleName
)

// GenesisState 创世状态
type GenesisState struct {
	// 简化的创世状态，主要包含基本配置
	Params EpochParams `json:"params"`
}

// DefaultGenesis 返回默认的创世状态
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultEpochParams(),
	}
}

// Validate 验证创世状态
func (gs GenesisState) Validate() error {
	// 验证参数
	if gs.Params.Kappa <= 0 || gs.Params.Kappa > 1 {
		return fmt.Errorf("invalid kappa: must be between 0 and 1")
	}
	if gs.Params.Alpha <= 0 || gs.Params.Alpha > 1 {
		return fmt.Errorf("invalid alpha: must be between 0 and 1")
	}
	if gs.Params.Delta <= 0 {
		return fmt.Errorf("invalid delta: must be positive")
	}

	return nil
}

// RegisterLegacyAminoCodec 注册 legacy amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// 注册消息类型
	cdc.RegisterConcrete(&MsgRunEpoch{}, "yuma/RunEpoch", nil)
}

// RegisterInterfaces 注册接口
func RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	// 注册消息类型
	reg.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRunEpoch{},
	)

	// 注册查询服务
	reg.RegisterImplementations((*grpc.Server)(nil),
		&QueryServer{},
	)
}

// RegisterQueryHandlerClient 注册查询处理器客户端
func RegisterQueryHandlerClient(ctx context.Context, mux *runtime.ServeMux, client QueryClient) error {
	// 注册查询路由
	mux.Handle("GET", pattern_Query_Epoch_0, func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_Query_Epoch_0(rctx, inboundMarshaler, client, req, params)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		forward_Query_Epoch_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)
	})
	return nil
}
