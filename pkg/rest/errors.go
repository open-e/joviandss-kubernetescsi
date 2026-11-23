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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	jcom "github.com/open-e/joviandss-kubernetescsi/pkg/common"
)

// The codes that say what an error means for the call that got it. They are
// the plugin's own, all below 1000, and are what GetCode returns and what
// the callers compare with.
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
	RestErrorResourceDNEShare               = 16
	RestErrorResourceDNEPool                = 17 // the pool is missing
	RestErrorResourceDNELun                 = 18 // the volume is not attached to the target
	RestErrorAuthFailed                     = 19 // the appliance refused the login: status 401 or 403
)

// A RestError carries two codes. GetCode says what the error means for the
// call, and is what the callers compare with. GetKindCode says which error
// the appliance sent: the code of its kind inside the range of its layer,
// the code of the layer alone for an error that no entry of the call holds,
// and 0 for a failure that is not an error of the appliance.
type RestError interface {
	Error() (out string)
	GetCode() int
	GetKindCode() int
}

type restError struct {
	code int
	kind int
	msg  string
}

func ErrCode(err RestError) int {
	if err != nil {
		return err.GetCode()
	}
	return RestErrorOk
}

// GetError makes the error of a failure that is not an error of the
// appliance: its second code is 0.
func GetError(c int, m string) RestError {
	return &restError{code: c, msg: m}
}

// applianceError makes the error of a reply of the appliance.
func applianceError(c int, kind int, m string) RestError {
	return &restError{code: c, kind: kind, msg: m}
}

func (err *restError) GetCode() int {
	return err.code
}

func (err *restError) GetKindCode() int {
	return err.kind
}

// codeText is what Error puts before the message of an error, by its code.
var codeText = map[int]string{
	RestErrorResourceBusy:                   "Resource is busy.",
	RestErrorResourceExists:                 "Object exists:",
	RestErrorRequestMalfunction:             "Failure in sending data to storage:",
	RestErrorResourceDNE:                    "Resource does not exist:",
	RestErrorUnableToConnect:                "Unable to connect to storage:",
	RestErrorRPM:                            "Failure during processing response from storage:",
	RestErrorStorageFailureUnknown:          "Storage failes with unknown error:",
	RestErrorRequestTimeout:                 "Request to storage timed out:",
	RestErrorArgumentIncorrect:              "Incorrect argument:",
	RestErrorResourceBusySnapshotHasClones:  "Snapshot has clones:",
	RestErrorResourceBusyVolumeHasSnapshots: "Volume has snapshots:",
	RestErrorOutOfSpace:                     "Out of space:",
	RestErrorResourceDNEVolume:              "Volume does not exist:",
	RestErrorResourceDNETarget:              "Target does not exist:",
	RestErrorResourceDNEShare:               "Share does not exist:",
	RestErrorResourceDNEPool:                "Pool does not exist:",
	RestErrorResourceDNELun:                 "Volume is not attached to target:",
	RestErrorAuthFailed:                     "Authentication failed:",
}

func (err *restError) Error() (out string) {
	if text, ok := codeText[err.code]; ok {
		return text + " " + err.msg
	}
	return "Unknown internal Error. " + err.msg
}

///////////////////////////////////////////////////////////////////////////////
// The layers and the kinds of the appliance's errors

// The layers of the appliance's errors, each with a range of 1000 codes.
// The class of an error shows which layer raised it. The code of a layer is
// the first of its range; its kinds of error take the numbers after it
// (error_catalogue.go).
const (
	LayerNoClass = 1000 // no class: login errors and some 400 and 404
	LayerPython  = 2000 // builtins.*: errors the appliance did not catch
	LayerZfs     = 3000 // zfslib.*: the ZFS wrapper
	LayerProduct = 4000 // opene.*: the product's own logic
	LayerSchema  = 5000 // jsonschema.*: a request body that fails its schema
	LayerRouting = 6000 // werkzeug.exceptions.*: routing and the API's own checks
)

// layerOf gives the code of the layer that the class of an error shows, and
// 0 for a class of no layer known.
func layerOf(class string) int {
	switch {
	case class == "":
		return LayerNoClass
	case strings.HasPrefix(class, "builtins."):
		return LayerPython
	case strings.HasPrefix(class, "zfslib."):
		return LayerZfs
	case strings.HasPrefix(class, "opene."):
		return LayerProduct
	case strings.HasPrefix(class, "jsonschema."):
		return LayerSchema
	case strings.HasPrefix(class, "werkzeug.exceptions."):
		return LayerRouting
	}
	return 0
}

// errorKind is one error model of the spec: how a reply of that kind looks.
type errorKind struct {
	kind    int
	status  int
	class   string // "" where the spec lists the class as absent
	errno   errnoField
	message *regexp.Regexp // nil where the spec gives the kind no pattern
}

// means makes the entry of a call's catalogue for this kind of error: code
// is what the error means for that call.
func (k errorKind) means(code int) errorEntry {
	return errorEntry{status: k.status, class: k.class, errno: k.errno, message: k.message, code: code, kind: k.kind}
}

// What an entry that is not in the spec puts where no recording gives the
// status or the class of its reply.
const (
	anyStatus = 0
	anyClass  = "*"
)

// errorEntry is one error a call can get, transcribed from the call's
// -scenarios.json file. A reply matches it on the status, the class, the
// errno in its three cases and the message pattern of the error's schema.
type errorEntry struct {
	status  int
	class   string
	errno   errnoField
	message *regexp.Regexp // nil: any message
	code    int            // what the error means for this call
	kind    int            // which error the appliance sent; 0: not in the spec
}

func (en errorEntry) matches(status int, e *ErrorT) bool {
	if en.status != anyStatus && en.status != status {
		return false
	}
	if en.class != anyClass && en.class != e.class() {
		return false
	}
	if en.errno != e.Errno {
		return false
	}
	if en.message == nil {
		return true
	}
	return e.Message != nil && en.message.MatchString(*e.Message)
}

///////////////////////////////////////////////////////////////////////////////
// The one handler of a failed call

// failedReply maps a failure of a call to one RestError and writes its one
// log line in pkg/rest: at Debug for an error the call's catalogue knows, at
// Error for an entry that says Unknown on purpose and for anything else.
// err is an error found before the reply was read as a failure: no reply, or
// a reply that does not decode.
func (s *RestEndpoint) failedReply(ctx context.Context, method, path string, entries []errorEntry,
	status int, body []byte, err RestError,
) RestError {
	l := jcom.LFC(ctx).WithFields(logrus.Fields{"section": "rest", "method": method, "path": path, "status": status})
	// the log line of a failure holds both of its codes
	withCodes := func(re RestError) *logrus.Entry {
		return l.WithFields(logrus.Fields{"code": re.GetCode(), "kind": re.GetKindCode()})
	}
	if err != nil {
		withCodes(err).Error(err.Error())
		return err
	}

	var reply struct {
		Error *ErrorT `json:"error"`
	}
	if json.Unmarshal(body, &reply) != nil || reply.Error == nil {
		// not a JSON error body: the status alone gives the code
		err = statusError(method, path, status, body)
		withCodes(err).Error(err.Error())
		return err
	}

	e := reply.Error
	l = l.WithFields(logrus.Fields{"class": e.class(), "errno": e.Errno.String(), "message": e.message()})
	for _, en := range entries { // the first entry that matches decides
		if en.matches(status, e) {
			kind := en.kind
			if kind == 0 {
				// not in the spec: the code of the layer alone
				kind = layerOf(e.class())
			}
			err = applianceError(en.code, kind, oneLine(e.message()))
			if en.code == RestErrorFailureUnknown {
				withCodes(err).Error("error of the catalogue, unknown on purpose")
			} else {
				withCodes(err).Debug("error of the catalogue")
			}
			return err
		}
	}
	code := RestErrorFailureUnknown
	if status == 404 {
		code = RestErrorResourceDNE // a 404 means not found, whatever its body
	}
	// no entry of the call holds it: its second code is that of its layer alone
	err = applianceError(code, layerOf(e.class()), fmt.Sprintf("%s %s: status %d, class %q, errno %s, message %q",
		method, path, status, e.class(), e.Errno, e.message()))
	withCodes(err).Error("error not in the catalogue of the call")
	return err
}

// statusError maps a reply without a JSON error body by its status alone,
// and keeps the first bytes of the body in the message. Such a reply has no
// class, so its second code is that of the layer without a class.
func statusError(method, path string, status int, body []byte) RestError {
	if len(body) > 200 {
		body = body[:200]
	}
	switch status {
	case 401, 403:
		return applianceError(RestErrorAuthFailed, LayerNoClass,
			fmt.Sprintf("%s %s: authentication failed, status %d: %q", method, path, status, body))
	case 404:
		return applianceError(RestErrorResourceDNE, LayerNoClass,
			fmt.Sprintf("%s %s: not found, status 404: %q", method, path, body))
	default:
		return applianceError(RestErrorFailureUnknown, LayerNoClass,
			fmt.Sprintf("%s %s: status %d without an error body: %q", method, path, status, body))
	}
}

// oneLine puts a message of the appliance on one line: some list the
// datasets in the way on lines of their own.
func oneLine(msg string) string {
	return strings.Join(strings.Fields(msg), " ")
}
