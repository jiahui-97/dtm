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
	"github.com/dtm-labs/dtm/examples"
	"github.com/stretchr/testify/assert"
)

func TestMsgNormal(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Submit()
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	waitTransProcessed(msg.Gid)
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgTimeoutSuccess(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.TransInResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusSubmitted, getTransStatus(msg.Gid))
	cronTransOnce()
	assert.Equal(t, []string{StatusSucceed, StatusSucceed}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusSucceed, getTransStatus(msg.Gid))
}

func TestMsgTimeoutFailed(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultOngoing)
	cronTransOnceForwardNow(180)
	assert.Equal(t, StatusPrepared, getTransStatus(msg.Gid))
	examples.MainSwitch.CanSubmitResult.SetOnce(dtmcli.ResultFailure)
	cronTransOnceForwardNow(180)
	assert.Equal(t, []string{StatusPrepared, StatusPrepared}, getBranchesStatus(msg.Gid))
	assert.Equal(t, StatusFailed, getTransStatus(msg.Gid))
}

func TestMsgAbnormal(t *testing.T) {
	msg := genMsg(dtmimp.GetFuncName())
	msg.Prepare("")
	err := msg.Prepare("")
	assert.Nil(t, err)
	err = msg.Submit()
	assert.Nil(t, err)

	err = msg.Prepare("")
	assert.Error(t, err)
}

func genMsg(gid string) *dtmcli.Msg {
	req := examples.GenTransReq(30, false, false)
	msg := dtmcli.NewMsg(examples.DtmHttpServer, gid).
		Add(examples.Busi+"/TransOut", &req).
		Add(examples.Busi+"/TransIn", &req)
	msg.QueryPrepared = examples.Busi + "/CanSubmit"
	return msg
}
