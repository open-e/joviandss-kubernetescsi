/*
Copyright (c) 2019 Open-E, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License. You may obtain
a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations
under the License.
*/

package rest

import (
	"context"
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
)

const (
	RestFailureUnknown		= 1
	RestResourceBusy		= 2
	RestResourceExists		= 3
	RestRequestMalfunction		= 4
	RestResourceDNE			= 5
	RestUnableToConnect		= 6
	RestRPM				= 7 // Response Processing Malfunction
	RestStorageFailureUnknown	= 8
	RestRequestTimeout		= 9
)

type RestError interface {
	Error() (out string)
	GetCode() int
}

type restError struct {
	code int
	msg  string
}

//TODO: Refactor to move logging of error message in this func
func GetError(c int, m string) RestError {
	out := restError{
		code: c,
		msg:  m,
	}
	return &out
}

const (
	resourceExistsMsgPattern = `Resource .*\/.* already exists\.`
	cloneCreateFailureDatasetExistsPattern = `cannot create .*\/.*: dataset already exists`
	// resourceExistsMsgPattern = `.*`
	resourceIsBusyMsgPattern = `In order to delete a zvol, you must delete all of its clones first.`
	resourceDneMsgPattern = `Zfs resource: (.+\/.+) not found in this collection`
	resourceHasClonesMsgPattern = `^In order to delete a zvol, you must delete all of its clones first\.$`
	resourceHasSnapshotsMsgPattern = `^cannot destroy '.*\/.*': volume has children.use '-r' to destroy the following datasets:\n.*`
	resourceHasClonesClassPattern = `^opene.storage.zfs.ZfsOeError$`
	resourceHasSnapshotsClassPattern = `^zfslib.wrap.zfs.ZfsCmdError$`
)

var resourceExistsMsgRegexp = regexp.MustCompile(resourceExistsMsgPattern)
var cloneCreateFailureDatasetExistsRegexp = regexp.MustCompile(cloneCreateFailureDatasetExistsPattern)
var resourceIsBusyMsgRegexp = regexp.MustCompile(resourceIsBusyMsgPattern)
var resourceDneMsgRegexp = regexp.MustCompile(resourceDneMsgPattern)
var resourceHasClonesMsgRegexp = regexp.MustCompile(resourceHasClonesMsgPattern)
var resourceHasSnapshotsMsgRegexp = regexp.MustCompile(resourceHasSnapshotsMsgPattern)
var resourceHasClonesClassRegexp = regexp.MustCompile(resourceHasClonesClassPattern)
var resourceHasSnapshotsClassRegexp = regexp.MustCompile(resourceHasSnapshotsClassPattern)


func ErrorFromErrorT(ctx context.Context, err *ErrorT, le *logrus.Entry) *restError {

	l := le.WithFields(logrus.Fields{
		"func": "ErrorFromErrorT",
		"traceId": ctx.Value("traceId"),
	})

	//if err, ok := errC.(*ErrorT); ok {
	//	return &restError{code: RestFailureUnknown, msg: *errC.Message}
	//}

	l.Debugf("ErrorT data %+v", err)
		if err.Errno != nil {
		if *err.Errno == 1 {
			if err.Message != nil {
				if resourceDneMsgRegexp.MatchString(*err.Message) {
					return &restError { code: RestResourceDNE }
				}
				if resourceExistsMsgRegexp.MatchString(*err.Message) {
					return &restError { code: RestResourceExists }
				}
			}

		}
		if *err.Errno == 5 {
			if err.Message != nil {
				l.Debug("Error 5")
				if resourceExistsMsgRegexp.MatchString(*err.Message) {
					l.Debug("Res exists!")

					return &restError { code: RestResourceExists }
				}
			}

		}
		if *err.Errno == 100 {
			if err.Message != nil {
				l.Debug("Error 5")
				if resourceExistsMsgRegexp.MatchString(*err.Message) {
					l.Debug("Resource exists!")
					return &restError { code: RestResourceExists }
				} else if cloneCreateFailureDatasetExistsRegexp.MatchString(*err.Message) {
					l.Debug("Clone exists!")
					return &restError { code: RestResourceExists }
				}
			}

		}

		if *err.Errno == 500 {
			if err.Message != nil {
				if resourceIsBusyMsgRegexp.MatchString(*err.Message) {
					return &restError { code: RestResourceBusy }
				} else if resourceDneMsgRegexp.MatchString(*err.Message) {
					match := resourceDneMsgRegexp.FindStringSubmatch(*err.Message)
					if len(match) > 1 {
						msg := fmt.Sprintf("Resource %s not found", match[1])
						l.Warnf(msg)
						return &restError { code: RestResourceBusy, msg: msg }
					}
					l.Warn("Resource not found")
					return &restError { code: RestResourceBusy, msg: *err.Message }
				}

			}
		}
	}
	l.Warn(err.String())
	//l.Warnf("Errno:%d, Class:%s, Message:%s, Url:%s", *err.Errno, *err.Class, *err.Message, *err.Url )
	return &restError{code: RestFailureUnknown}
}

func (err *restError) Error() (out string) {

	switch (*err).code {

	case RestResourceBusy:
		out = fmt.Sprintf("Resource is busy. %s", err.msg)

	case RestRequestMalfunction:
		out = fmt.Sprintf("Failure in sending data to storage: %s", err.msg)

	case RestRPM:
		out = fmt.Sprintf("Failure during processing response from storage: %s", err.msg)
	case RestResourceDNE:
		out = fmt.Sprintf("Resource %s do not exists", err.msg)
	case RestResourceExists:
		out = fmt.Sprintf("Object exists: %s", err.msg)

	case RestStorageFailureUnknown:
		out = fmt.Sprintf("Storage failes with unknown error: %s", err.msg)

	default:
		out = fmt.Sprint("Unknown internal Error. %s", err.msg)

	}
	return out
}

func (err *restError) GetCode() int {
	return err.code

}
