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
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	RestErrorOk                             = 0
	RestErrorFailureUnknown                 = 1
	RestErrorResourceBusy                   = 2
	RestErrorResourceExists                 = 3
	RestErrorRequestMalfunction             = 4
	RestErrorResourceDNE                    = 5
	RestErrorUnableToConnect                = 6
	RestErrorRPM                            = 7 // Response Processing Malfunction
	RestErrorStorageFailureUnknown          = 8
	RestErrorRequestTimeout                 = 9
	RestErrorArgumentIncorrect              = 10
	RestErrorResourceBusySnapshotHasClones  = 11
	RestErrorResourceBusyVolumeHasSnapshots = 12
	RestErrorOutOfSpace                     = 13
	RestErrorResourceDNEVolume              = 14
	RestErrorResourceDNETarget              = 15
)

type RestError interface {
	Error() (out string)
	GetCode() int
}

type restError struct {
	code int
	msg  string
}

func ErrCode(err RestError) int {
	if err != nil {
		return err.GetCode()
	}
	return RestErrorOk
}

// TODO: Refactor to move logging of error message in this func
func GetError(c int, m string) RestError {
	out := restError{
		code: c,
		msg:  m,
	}
	return &out
}

const (
	resourceExistsMsgPattern               = `Resource .*\/.* already exists\.`
	cloneCreateFailureDatasetExistsPattern = `cannot create .*\/.*: dataset already exists`
	// resourceExistsMsgPattern = `.*`
	resourceIsBusyMsgPattern    = `In order to delete a zvol, you must delete all of its clones first.`
	resourceDneMsgPattern       = `Zfs resource: (.+\/.+) not found in this collection`
	volumeDneMsgPattern         = `Volume (?P<volume>[\w\-/\.]+) not found in pool (?P<pool>[\w\-/\.]+).`
	itemNotFoundClassPattern    = `opene.exceptions.ItemNotFoundError`
	snapshotDneMsgPatterm       = `cannot open '([\w\-\/\.]+@[\w\-\.]+)': dataset does not exist`
	resourceHasClonesMsgPattern = `^In order to delete a zvol, you must delete all of its clones first\.$`
	volumeHasChildrenMsgPattern = `^cannot destroy '(?P<volume>[\w\-/\.]+)': volume has children[\s\S]use '-r' to destroy the following datasets:(?P<datasets>[.\s\S]*)`
	snapshotHasClonesMsgPattern = `^cannot destroy '(?P<snapshot>[\w\-/\.]+@[\w\-\.]+)': snapshot has dependent clones[\s\S]use '-R' to destroy the following datasets(?P<datasets>[.\s\S]*)`
	//`^cannot destroy '([\w\-\/]+@[\w\-]+)': snapshot has dependent clones\nuse '-R' to destroy the following datasets:(.*)`
	resourceHasClonesClassPattern    = `^opene.storage.zfs.ZfsOeError$`
	resourceHasSnapshotsClassPattern = `^zfslib.wrap.zfs.ZfsCmdError$`
	zfsCmdErrorPattern               = `^zfslib.wrap.zfs.ZfsCmdError$`
	storageResourceExhaustedPattern  = `New zvol size\(([\d]+)\) exceeds available space on pool ([\w\-\.]+)\(([\d]+)\).`
	targetNameConflictClassPattern   = `opene.san.target.base.iscsi.target.TargetNameConflictError`
	targetExistsMsgPattern           = `Target with name (?P<target>[a-z0-9\.:\-]*) is already present on`
	targetDneMsgPattern              = `Target iqn.20215:test2asd not exists.`
	lunIdUsedMsgPattern              = `LUN ([\d]+) is already used in (?P<target>[a-z0-9\.:\-]*).`
	volumeAlreadyUsedPattern         = `Volume (?P<volumesyspath>.*) is already used.`
	lunItemConflictClassPattern      = `opene.exceptions.ItemConflictError`
)

var resourceExistsMsgRegexp = regexp.MustCompile(resourceExistsMsgPattern)
var cloneCreateFailureDatasetExistsRegexp = regexp.MustCompile(cloneCreateFailureDatasetExistsPattern)
var resourceIsBusyMsgRegexp = regexp.MustCompile(resourceIsBusyMsgPattern)
var resourceDneMsgRegexp = regexp.MustCompile(resourceDneMsgPattern)
var volumeDneMsgRegexp = regexp.MustCompile(volumeDneMsgPattern)
var itemNotFoundClassRegexp = regexp.MustCompile(itemNotFoundClassPattern)
var snapshotDneMsgRegexp = regexp.MustCompile(snapshotDneMsgPatterm)
var resourceHasClonesMsgRegexp = regexp.MustCompile(resourceHasClonesMsgPattern)
var volumeHasChildrenMsgRegexp = regexp.MustCompile(volumeHasChildrenMsgPattern)
var snapshotHasClonesMsgRegexp = regexp.MustCompile(snapshotHasClonesMsgPattern)
var resourceHasClonesClassRegexp = regexp.MustCompile(resourceHasClonesClassPattern)
var resourceHasSnapshotsClassRegexp = regexp.MustCompile(resourceHasSnapshotsClassPattern)
var zfsCmdErrorRegexp = regexp.MustCompile(zfsCmdErrorPattern)
var storageResourceExhaustedRegexp = regexp.MustCompile(storageResourceExhaustedPattern)
var targetNameConflictClassRegexp = regexp.MustCompile(targetNameConflictClassPattern)
var targetExistsMsgRegexp = regexp.MustCompile(targetExistsMsgPattern)
var targetDneMsgRegexp = regexp.MustCompile(targetDneMsgPattern)

var lunIdUsedMsgRegexp = regexp.MustCompile(lunIdUsedMsgPattern)
var volumeAlteadyUsedRegexp = regexp.MustCompile(volumeAlreadyUsedPattern)
var lunItemConflictClassRegexp = regexp.MustCompile(lunItemConflictClassPattern)

func ErrorFromErrorT(ctx context.Context, err *ErrorT, le *logrus.Entry) *restError {

	l := le.WithFields(logrus.Fields{
		"func":    "ErrorFromErrorT",
		"section": "rest",
	})

	if err == nil {
		l.Warn("Code is trying to get Error for No error case")

		stackTrace := debug.Stack()
		le.Warnln("Manual stack trace log:")
		le.Warnln(string(stackTrace))
	}

	l.Debugf("ErrorT data %+v", err)
	if err.Errno != nil {
		l.Warnln("Error number ", *err.Errno)

		switch *err.Errno {

		case 0:
			if err.Message != nil {
				// Check if that is DNE message
				if snapshotDneMsgRegexp.MatchString(*err.Message) {
					if err.Class != nil {
						if zfsCmdErrorRegexp.MatchString(*err.Class) {
							return &restError{code: RestErrorResourceDNE}
						}
					}
					return &restError{code: RestErrorResourceDNE}
				}
			}

		case 1:
			if err.Message != nil {
				if resourceDneMsgRegexp.MatchString(*err.Message) {
					match := resourceDneMsgRegexp.FindStringSubmatch(*err.Message)
					if len(match) > 1 {
						msg := fmt.Sprintf("Resource %s not found", match[1])
						l.Warnf(msg)
						return &restError{code: RestErrorResourceDNE, msg: msg}
					}
					l.Warn("Resource not found")
					return &restError{code: RestErrorResourceDNE, msg: *err.Message}
				}
				if resourceExistsMsgRegexp.MatchString(*err.Message) {
					return &restError{code: RestErrorResourceExists}
				}
			}

		case 5:
			if err.Message != nil {
				if resourceExistsMsgRegexp.MatchString(*err.Message) {
					l.Debug("Res exists!")

					return &restError{code: RestErrorResourceExists}
				}
			}
		case 100:
			if err.Message != nil {
				if resourceExistsMsgRegexp.MatchString(*err.Message) {
					l.Debug("Resource exists!")
					return &restError{code: RestErrorResourceExists}
				} else if cloneCreateFailureDatasetExistsRegexp.MatchString(*err.Message) {
					l.Debug("Clone exists!")
					return &restError{code: RestErrorResourceExists}
				}
			}
		case 1000:

			if err.Message != nil {
				if snapshotHasClonesMsgRegexp.MatchString(*err.Message) {
					match := snapshotHasClonesMsgRegexp.FindStringSubmatch(*err.Message)
					datasets := snapshotHasClonesMsgRegexp.SubexpIndex("datasets")
					snapshot := snapshotHasClonesMsgRegexp.SubexpIndex("snapshot")

					msg := fmt.Sprintf("Snapshot %s has dependent resources %s", match[snapshot], strings.Replace(match[datasets], "\n", " ", -1))
					l.Debug(msg)
					return &restError{code: RestErrorResourceBusySnapshotHasClones, msg: msg}
				}
				if volumeHasChildrenMsgRegexp.MatchString(*err.Message) {
					match := volumeHasChildrenMsgRegexp.FindStringSubmatch(*err.Message)
					datasets := volumeHasChildrenMsgRegexp.SubexpIndex("datasets")
					volume := volumeHasChildrenMsgRegexp.SubexpIndex("volume")

					msg := fmt.Sprintf("Volume %s has dependent resources %s", match[volume], strings.Replace(match[datasets], "\n", " ", -1))
					l.Debug(msg)
					return &restError{code: RestErrorResourceBusyVolumeHasSnapshots, msg: msg}
				}
			}
		}
	} else {
		if err.Class != nil {
			if err.Message != nil {
				if targetNameConflictClassRegexp.MatchString(*err.Class) && targetExistsMsgRegexp.MatchString(*err.Message) {
					return &restError{code: RestErrorResourceExists}
				}
				if lunItemConflictClassRegexp.MatchString(*err.Class) {
					if lunIdUsedMsgRegexp.MatchString(*err.Message) {
						return &restError{code: RestErrorResourceExists, msg: *err.Message}
					}
					if volumeAlteadyUsedRegexp.MatchString(*err.Message) {
						return &restError{code: RestErrorResourceExists, msg: *err.Message}
					}
				}
				if itemNotFoundClassRegexp.MatchString(*err.Class) {
					if volumeDneMsgRegexp.MatchString(*err.Message) {
						return &restError{code: RestErrorResourceDNEVolume, msg: *err.Message}
					}
					if targetDneMsgRegexp.MatchString(*err.Message) {
						return &restError{code: RestErrorResourceDNETarget, msg: *err.Message}
					}
				}
			}
		}
		if err.Message != nil {
			if resourceIsBusyMsgRegexp.MatchString(*err.Message) {
				return &restError{code: RestErrorResourceBusy}
			} else if resourceDneMsgRegexp.MatchString(*err.Message) {
				match := resourceDneMsgRegexp.FindStringSubmatch(*err.Message)
				if len(match) > 1 {
					msg := fmt.Sprintf("Resource %s not found", match[1])
					l.Warnf(msg)
					return &restError{code: RestErrorResourceDNE, msg: msg}
				}
				l.Warn("Resource not found")
				return &restError{code: RestErrorResourceDNE, msg: *err.Message}
			} else if storageResourceExhaustedRegexp.MatchString(*err.Message) {
				return &restError{code: RestErrorOutOfSpace, msg: *err.Message}
			}
		}
	}
	l.Warnln("Unable to identify error: ", err.String())
	//l.Warnf("Errno:%d, Class:%s, Message:%s, Url:%s", *err.Errno, *err.Class, *err.Message, *err.Url )
	return &restError{code: RestErrorFailureUnknown}
}

func (err *restError) Error() (out string) {

	switch (*err).code {

	case RestErrorResourceBusy:
		out = fmt.Sprintf("Resource is busy. %s", err.msg)

	case RestErrorRequestMalfunction:
		out = fmt.Sprintf("Failure in sending data to storage: %s", err.msg)
	case RestErrorRPM:
		out = fmt.Sprintf("Failure during processing response from storage: %s", err.msg)
	case RestErrorResourceDNE:
		out = fmt.Sprintf("Resource %s do not exists", err.msg)
	case RestErrorResourceExists:
		out = fmt.Sprintf("Object exists: %s", err.msg)
	case RestErrorStorageFailureUnknown:
		out = fmt.Sprintf("Storage failes with unknown error: %s", err.msg)

	default:
		out = fmt.Sprintf("Unknown internal Error. %s", err.msg)

	}
	return out
}

func (err *restError) GetCode() int {
	return err.code

}
