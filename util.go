package dlog

import (
	"context"
	"fmt"
	"github.com/dajinkuang/util/ordermaputil"
)

const __ContextDLogOrderMapKey = "context_order_map_key"

// SetTraceInfo 设置trace信息
func SetTraceInfo(ctx context.Context, traceID, parentID, spanID string) context.Context {
	om := ordermaputil.NewOrderMap()
	om.Set(TraceID, traceID)
	om.Set(ParentID, parentID)
	om.Set(SpanID, spanID)
	src := FromContext(ctx)
	if src == nil {
		src = ordermaputil.NewOrderMap()
	}
	src.AddValues(om)
	return setContext(ctx, src)
}

// CopyTraceInfo 拷贝trace信息其它的全部丢弃，比如超时设置等
func CopyTraceInfo(ctx context.Context) context.Context {
	src := FromContext(ctx)
	if src == nil {
		src = ordermaputil.NewOrderMap()
	}
	return setContext(context.Background(), src)
}

// GetTraceInfo 获取trace信息
func GetTraceInfo(ctx context.Context) (traceID, parentID, spanID string) {
	om := FromContext(ctx)
	if tmp, ok := om.Get(TraceID); ok {
		traceID = tmp.(string)
	}
	if tmp, ok := om.Get(ParentID); ok {
		parentID = tmp.(string)
	}
	if tmp, ok := om.Get(SpanID); ok {
		spanID = tmp.(string)
	}
	return
}

// FromContext 获取ctx中存的OrderMap
func FromContext(ctx context.Context) *ordermaputil.OrderMap {
	ret := ctx.Value(__ContextDLogOrderMapKey)
	if ret == nil {
		return nil
	}
	return ret.(*ordermaputil.OrderMap)
}

// setContext 将OrderMap设置到ctx中
func setContext(ctx context.Context, dt *ordermaputil.OrderMap) context.Context {
	ctx = context.WithValue(ctx, __ContextDLogOrderMapKey, dt)
	return ctx
}

// ValueFromOM 从ctx中根据key获取值
func ValueFromOM(ctx context.Context, key interface{}) interface{} {
	src := FromContext(ctx)
	if src == nil {
		return nil
	}
	val, ok := src.Get(fmt.Sprintf("%v", key))
	if !ok {
		return nil
	}
	return val
}
