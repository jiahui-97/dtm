/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"

	"github.com/dtm-labs/dtm/dtmcli"
	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmgrpc"
	"github.com/dtm-labs/dtm/examples"
	"github.com/stretchr/testify/assert"
)

func TestSagaGrpcNormal(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaGrpcRollback(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, true)
	examples.MainSwitch.TransOutRevertResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusAborting, getTransStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, StatusFailed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusSucceed, StatusSucceed, StatusSucceed, StatusFailed}, getBranchesStatus(saga.Gid))
}

func TestSagaGrpcCurrent(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false).
		EnableConcurrent()
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaGrpcCurrentOrder(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false).
		EnableConcurrent().
		AddBranchOrder(1, []int{0})
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func TestSagaGrpcCommittedOngoing(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false)
	examples.MainSwitch.TransOutResult.SetOnce(dtmcli.ResultOngoing)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, StatusSubmitted, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusPrepared, StatusPrepared, StatusPrepared}, getBranchesStatus(saga.Gid))
	cronTransOnce()
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
}

func TestSagaGrpcNormalWait(t *testing.T) {
	saga := genSagaGrpc(dtmimp.GetFuncName(), false, false)
	saga.SetOptions(&dtmcli.TransOptions{WaitResult: true})
	saga.Submit()
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
	waitTransProcessed(saga.Gid)
}

func TestSagaGrpcEmptyUrl(t *testing.T) {
	saga := dtmgrpc.NewSagaGrpc(examples.DtmGrpcServer, dtmimp.GetFuncName())
	req := examples.GenBusiReq(30, false, false)
	saga.Add(examples.BusiGrpc+"/examples.Busi/TransOut", examples.BusiGrpc+"/examples.Busi/TransOutRevert", req)
	saga.Add("", examples.BusiGrpc+"/examples.Busi/TransInRevert", req)
	saga.Submit()
	waitTransProcessed(saga.Gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(saga.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(saga.Gid))
}

func genSagaGrpc(gid string, outFailed bool, inFailed bool) *dtmgrpc.SagaGrpc {
	saga := dtmgrpc.NewSagaGrpc(examples.DtmGrpcServer, gid)
	req := examples.GenBusiReq(30, outFailed, inFailed)
	saga.Add(examples.BusiGrpc+"/examples.Busi/TransOut", examples.BusiGrpc+"/examples.Busi/TransOutRevert", req)
	saga.Add(examples.BusiGrpc+"/examples.Busi/TransIn", examples.BusiGrpc+"/examples.Busi/TransInRevert", req)
	return saga
}
