/*
Copyright (c) 2026 Open-E, Inc.
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

// Package rest is the client of the JovianDSS REST API that the CSI plugin
// uses. It speaks API v4 of JovianDSS 1 only: every path starts with
// api/v4, and there is no setting for another version and no fallback.
//
// # Where the types come from
//
// The package follows what appliances were recorded to do, not what the
// documentation of the API says. The source is the jdss-rest-api-spec
// project, which records the replies of real appliances and describes them:
// an OpenAPI of the observed replies, and for each call a scenarios file of
// the errors recorded for it.
//
// The reply types (types.go) are transcribed from that OpenAPI, and the
// header of types.go names the commit of jdss-rest-api-spec they were read
// from. The catalogue of errors (error_catalogue.go) is transcribed from
// the scenarios files, and its header names the commit too.
//
// # Read fields are exact, the others are raw JSON
//
// One field with the wrong JSON type makes Go's decoder fail the whole
// reply. So only the fields the plugin reads carry a type, the exact type
// the recordings show for that endpoint, and every other field of a reply
// is kept as raw JSON, which no type of the appliance can fail.
//
// A read field is required or optional, as the recordings show it. A reply
// that lacks a required field fails with an error that names the field,
// "required reply field <name> is missing"; it is never read as zero. A
// number or a date that the appliance sends as a string is converted when
// the reply is decoded, and a value that does not parse fails the reply
// with "reply field <name> does not convert". A reply without data, or
// with data of another shape, is an error, never an empty object or an
// empty list: a caller would read those as a volume of size 0, a full pool
// or a snapshot without clones.
//
// # Errors
//
// Every call ends in its success or in one function, failedReply, with the
// catalogue of the errors that call can get. The status of a reply is read
// first. A reply matches an entry of the catalogue on its status, class,
// errno and message pattern; an errno that is a number, one that is null
// and one that is missing are three different inputs.
//
// A RestError carries two codes. GetCode says what the error means for the
// call, and is what a caller compares with: not found, exists, busy, out of
// space and so on. GetKindCode says which error the appliance sent: each
// layer of the appliance, shown by the class of the error, has a range of
// 1000 codes, and a kind of error has its number inside the range of its
// layer; an error that no entry of the call holds has the code of its layer
// alone, and a failure that is not an error of the appliance has 0.
//
// An error that matches no entry of its call is RestErrorFailureUnknown,
// with its status, class, errno and message kept in the error, unless its
// status gives the meaning: a 404 is "not found" whatever its body, and a
// 401 or 403 without a JSON body is a failed authentication. Nothing else
// is guessed. A caller decides on the code alone, never on message text.
//
// # Logging
//
// A request is logged as one line at Debug: the method, the path and the
// status. No body and no header of a request is logged at any level: a
// body can hold a CHAP password, and the Authorization header holds the
// REST password. A failed call is logged once in this package, by
// failedReply: at Debug when the catalogue knows the error, at Error
// otherwise.
//
// # Keeping the package in step with the appliance
//
// The checks are repeatable. The unit tests decode recorded replies kept
// in testdata, pass every recorded error through the handler, and serve
// both to the calls from a fake appliance; run them with "make test". When
// jdss-rest-api-spec gains recordings, read the types and the catalogue
// again from the files that changed, name the new commit in both headers,
// refresh the fixtures, and run "make test" and the audit of the types
// against the spec again.
package rest
