/*
Copyright (c) 2024 Open-E, Inc.
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

package common

// Plugin name
// This variable should be consistent with value that is used in StorageClass deffinition
var PluginName = "iscsi.csi.joviandss.open-e.com"

// CLI Protocol Type variable used in CLI
// Specifyes what protocol to use for providing volumes
// available iscsi and nfs
var CLIStorageAccessProtocolType StorageAccessProtocolType = "iscsi"

// Version of plugin
// Version gets initialized during compilation
var Version  string

// Node ID
// Stores Node Identifier for host that is running plugin
// Get initialized during plugin start through arguments
// If nothing is provided, app will try to determine value by scanning
// machine environment
var NodeID string

// Log level
var LogLevel string

// Log Path
// Where to store log file
var LogPath  string


// Controller Config Path
// Path to config file
var ControllerConfigPath string

