package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/ireward/wago/logger/tag"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_ParseZapLevel(t *testing.T) {
	assert.Equal(t, zap.DebugLevel, parseZapLevel("debug"))
	assert.Equal(t, zap.InfoLevel, parseZapLevel("info"))
	assert.Equal(t, zap.WarnLevel, parseZapLevel("warn"))
	assert.Equal(t, zap.ErrorLevel, parseZapLevel("error"))
	assert.Equal(t, zap.FatalLevel, parseZapLevel("fatal"))
	assert.Equal(t, zap.InfoLevel, parseZapLevel("unknown"))
}

func Test_DefaultZapLogger(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		assert.NoError(t, err)
		outC <- buf.String()
	}()

	logger := NewZapLogger(zap.NewExample())
	preCaller := caller(1)
	logger.With(tag.NewErrorTag(fmt.Errorf("test error"))).Info("test info", tag.NewStringTag("test", "field"))

	w.Close()
	os.Stdout = old
	out := <-outC
	sps := strings.Split(preCaller, ":")
	par, err := strconv.Atoi(sps[1])
	assert.Nil(t, err)
	lineNum := fmt.Sprintf("%v", par+1)
	assert.Equal(t, `{"level":"info","msg":"test info","error":"test error","test":"field","logging-call-at":"zap_logger_test.go:`+lineNum+`"}`+"\n", out)
}

func Test_CtxZapLogger(t *testing.T) {
	ctx := context.Background()
	logger := NewZapLogger(zap.NewExample())
	logger1 := FromCtx(ctx)
	// we always expect a logger
	assert.NotNil(t, logger1)

	ctx = WithCtx(ctx, logger)

	logger2 := FromCtx(ctx)
	assert.NotNil(t, logger2)

}
