package tag

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapTag is a wrapper over zap.Field
type ZapTag struct {
	// we keep this field private
	field zap.Field
}

func NewZapTag(field zap.Field) ZapTag {
	return ZapTag{
		field: field,
	}
}

func (t ZapTag) Field() zap.Field {
	return t.field
}

func (t ZapTag) Key() string {
	return t.field.Key
}

func (t ZapTag) Value() interface{} {
	enc := zapcore.NewMapObjectEncoder()
	t.field.AddTo(enc)
	for _, val := range enc.Fields {
		return val
	}
	return nil
}

func NewStringTag(key, value string) ZapTag {
	return ZapTag{
		field: zap.String(key, value),
	}
}

func NewStringsTag(key string, value []string) ZapTag {
	return ZapTag{
		field: zap.Strings(key, value),
	}
}

func NewDurationTag(key string, value time.Duration) ZapTag {
	return ZapTag{
		field: zap.Duration(key, value),
	}
}

func NewTimeTag(key string, value time.Time) ZapTag {
	return ZapTag{
		field: zap.Time(key, value),
	}
}

func NewErrorTag(val error) ZapTag {
	return ZapTag{
		field: zap.Error(val),
	}
}

func NewAnyTag(key string, value interface{}) ZapTag {
	return ZapTag{
		field: zap.Any(key, value),
	}
}

func NewBoolTag(key string, value bool) ZapTag {
	return ZapTag{
		field: zap.Bool(key, value),
	}
}

func NewIntTag(key string, value int) ZapTag {
	return ZapTag{
		field: zap.Int(key, value),
	}
}

func NewInt32sTag(key string, value []int32) ZapTag {
	return ZapTag{
		field: zap.Int32s(key, value),
	}
}

func NewInt32Tag(key string, value int32) ZapTag {
	return ZapTag{
		field: zap.Int32(key, value),
	}
}

func NewInt64Tag(key string, value int64) ZapTag {
	return ZapTag{
		field: zap.Int64(key, value),
	}
}
