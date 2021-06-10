/*
Copyright 2021 The Opensourceways Group.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package middleware

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
	"github.com/opensourceways/app-community-metadata/app"
	"go.uber.org/zap"
)

func RequestLog() gin.HandlerFunc {
	//skip health check endpoint
	skip := map[string]int{
		"/health": 1,
		"/status": 1,
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		reqId := strutil.Md5(fmt.Sprintf("%d", start.Nanosecond()))

		// add reqID to context
		c.Set("reqId", reqId)

		// c.MustBindWith()
		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; ok {
			return
		}

		// log post/put data
		postData := ""
		if c.Request.Method != "GET" {
			buf, _ := ioutil.ReadAll(c.Request.Body)
			postData = string(buf)
		}

		app.Logger.Info(
			"complete request",
			zap.String("req_id", reqId),
			zap.Namespace("context"),
			zap.String("req_date", start.Format("2006-01-02 15:04:05")),
			zap.String("method", c.Request.Method),
			zap.String("uri", c.Request.URL.String()),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("http_status", c.Writer.Status()),
			zap.String("elapsed_time", mathutil.ElapsedTime(start)),
			zap.String("post_data", postData),
		)
	}
}
