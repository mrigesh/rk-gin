// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginpanic

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

// PanicInterceptor returns a gin.HandlerFunc (middleware)
func Interceptor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if recv := recover(); recv != nil {
				var res *rkerror.ErrorResp

				if se, ok := recv.(*rkerror.ErrorResp); ok {
					res = se
				} else if re, ok := recv.(error); ok {
					res = rkerror.FromError(re)
				} else {
					res = rkerror.New(rkerror.WithMessage(fmt.Sprintf("%v", recv)))
				}

				rkginctx.GetEvent(ctx).SetCounter("panic", 1)
				rkginctx.GetEvent(ctx).AddErr(res.Err)
				rkginctx.GetLogger(ctx).Error(fmt.Sprintf("panic occurs:\n%s", string(debug.Stack())), zap.Error(res.Err))

				ctx.JSON(http.StatusInternalServerError, res)
			}
		}()

		ctx.Next()
	}
}
