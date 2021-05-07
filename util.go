package dlog

import (
	"context"
	"fmt"
)

const __CtxDLogOrderMapKey = "ctx_dlog_order_map_key"

func SetTraceInfo(ctx context.Context, traceID, parentID, spanID string) context.Context {
	om := NewOrderMap()
	om.Set(TraceID, traceID)
	om.Set(ParentID, parentID)
	om.Set(SpanID, spanID)
	src := FromContext(ctx)
	if src == nil {
		src = NewOrderMap()
	}
	src.AddVals(om)
	return setContext(ctx, src)
}

// 其它的全部丢弃，比如超时设置等
func CopyTraceInfo(ctx context.Context) context.Context {
	src := FromContext(ctx)
	if src == nil {
		src = NewOrderMap()
	}
	return setContext(context.Background(), src)
}

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

func FromContext(ctx context.Context) *OrderedMap {
	ret := ctx.Value(__CtxDLogOrderMapKey)
	if ret == nil {
		return nil
	}
	return ret.(*OrderedMap)
}

// 别人不需要用
func setContext(ctx context.Context, dt *OrderedMap) context.Context {
	ctx = context.WithValue(ctx, __CtxDLogOrderMapKey, dt)
	return ctx
}

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
