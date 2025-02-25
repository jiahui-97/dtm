/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmcli/logger"
	"github.com/dtm-labs/dtm/examples"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestTccBarrierNormal(t *testing.T) {
	req := examples.GenTransReq(30, false, false)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		assert.Nil(t, err)
		return tcc.CallBranch(req, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	assert.Nil(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
}

func TestTccBarrierRollback(t *testing.T) {
	req := examples.GenTransReq(30, false, true)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		_, err := tcc.CallBranch(req, Busi+"/TccBTransOutTry", Busi+"/TccBTransOutConfirm", Busi+"/TccBTransOutCancel")
		assert.Nil(t, err)
		return tcc.CallBranch(req, Busi+"/TccBTransInTry", Busi+"/TccBTransInConfirm", Busi+"/TccBTransInCancel")
	})
	assert.Error(t, err)
	waitTransProcessed(gid)
	assert.Equal(t, StatusFailed, getTransStatus(gid))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared, StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
}

var disorderHandler func(c *gin.Context) (interface{}, error) = nil

func TestTccBarrierDisorder(t *testing.T) {
	timeoutChan := make(chan string, 2)
	finishedChan := make(chan string, 2)
	gid := dtmimp.GetFuncName()
	err := dtmcli.TccGlobalTransaction(DtmServer, gid, func(tcc *dtmcli.Tcc) (*resty.Response, error) {
		body := &examples.TransReq{Amount: 30}
		tryURL := Busi + "/TccBTransOutTry"
		confirmURL := Busi + "/TccBTransOutConfirm"
		cancelURL := Busi + "/TccBSleepCancel"
		// 请参见子事务屏障里的时序图，这里为了模拟该时序图，手动拆解了callbranch
		branchID := tcc.NewSubBranchID()
		sleeped := false
		disorderHandler = func(c *gin.Context) (interface{}, error) {
			res, err := examples.TccBarrierTransOutCancel(c)
			if !sleeped {
				sleeped = true
				logger.Debugf("sleep before cancel return")
				<-timeoutChan
				finishedChan <- "1"
			}
			return res, err
		}
		// 注册子事务
		resp, err := dtmimp.RestyClient.R().
			SetBody(map[string]interface{}{
				"gid":                tcc.Gid,
				"branch_id":          branchID,
				"trans_type":         "tcc",
				"status":             StatusPrepared,
				"data":               string(dtmimp.MustMarshal(body)),
				dtmcli.BranchConfirm: confirmURL,
				dtmcli.BranchCancel:  cancelURL,
			}).Post(fmt.Sprintf("%s/%s", tcc.Dtm, "registerBranch"))
		assert.Nil(t, err)
		assert.Contains(t, resp.String(), dtmcli.ResultSuccess)

		go func() {
			logger.Debugf("sleeping to wait for tcc try timeout")
			<-timeoutChan
			r, _ := dtmimp.RestyClient.R().
				SetBody(body).
				SetQueryParams(map[string]string{
					"dtm":        tcc.Dtm,
					"gid":        tcc.Gid,
					"branch_id":  branchID,
					"trans_type": "tcc",
					"op":         dtmcli.BranchTry,
				}).
				Post(tryURL)
			assert.True(t, strings.Contains(r.String(), dtmcli.ResultSuccess)) // 这个是悬挂操作，为了简单起见，依旧让他返回成功
			finishedChan <- "1"
		}()
		logger.Debugf("cron to timeout and then call cancel")
		go cronTransOnceForwardNow(300)
		time.Sleep(100 * time.Millisecond)
		logger.Debugf("cron to timeout and then call cancelled twice")
		cronTransOnceForwardNow(300)
		timeoutChan <- "wake"
		timeoutChan <- "wake"
		<-finishedChan
		<-finishedChan
		time.Sleep(100 * time.Millisecond)
		return nil, fmt.Errorf("a cancelled tcc")
	})
	assert.Error(t, err, fmt.Errorf("a cancelled tcc"))
	assert.Equal(t, []string{StatusSucceed, StatusPrepared}, getBranchesStatus(gid))
	assert.Equal(t, StatusFailed, getTransStatus(gid))
}

func TestTccBarrierPanic(t *testing.T) {
	bb := &dtmcli.BranchBarrier{TransType: "saga", Gid: "gid1", BranchID: "bid1", Op: "action", BarrierID: 1}
	var err error
	func() {
		defer dtmimp.P2E(&err)
		tx, _ := dbGet().ToSQLDB().BeginTx(context.Background(), &sql.TxOptions{})
		bb.Call(tx, func(tx *sql.Tx) error {
			panic(fmt.Errorf("an error"))
		})
	}()
	assert.Error(t, err, fmt.Errorf("an error"))
}
