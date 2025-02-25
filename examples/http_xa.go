/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package examples

import (
	"github.com/dtm-labs/dtm/common"
	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

// XaClient XA client connection
var XaClient *dtmcli.XaClient = nil

func init() {
	setupFuncs["XaSetup"] = func(app *gin.Engine) {
		var err error
		XaClient, err = dtmcli.NewXaClient(DtmHttpServer, config.ExamplesDB, Busi+"/xa", func(path string, xa *dtmcli.XaClient) {
			app.POST(path, common.WrapHandler(func(c *gin.Context) (interface{}, error) {
				return xa.HandleCallback(c.Query("gid"), c.Query("branch_id"), c.Query("op"))
			}))
		})
		logger.FatalIfError(err)
	}
	addSample("xa", func() string {
		gid := dtmcli.MustGenGid(DtmHttpServer)
		err := XaClient.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
			resp, err := xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOutXa")
			if err != nil {
				return resp, err
			}
			return xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInXa")
		})
		logger.FatalIfError(err)
		return gid
	})
}
