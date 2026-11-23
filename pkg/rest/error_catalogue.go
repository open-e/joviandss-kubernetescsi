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

package rest

// The catalogue of the errors of the appliance: the kinds of error, and for
// each call of the plugin the errors that call can get.
//
// It is transcribed from the description of the recorded API in
// jdss-rest-api-spec, commit 434c9cd: the -scenarios.json file of each call,
// under empirical/jdss1/v4/typespec/<path>/<METHOD>-scenarios.json. Such a
// file gives for every error recorded for the call its status, class, errno,
// the pattern of its message and its kind, the error model of the spec. When
// that repository gains recordings, read the catalogue again from the
// scenarios files that changed since that commit, and name the new commit
// here. A kind keeps its number; a new kind takes the next free number of
// its layer (docs/adr/0002-api-v4-compliance-jdss1.md, Table 19).
//
// An entry gives two codes. The first is what the error means for the call,
// the code GetCode returns: it is the code the call returned for that reply
// before the catalogue, Unknown included, and another only where the design
// record decides one; those entries carry a comment. The second is the kind.
// The entries of a call are tried in order and the first match decides, so
// an entry whose pattern also matches the message of another stands after
// that other.

import "regexp"

// The kinds of error: one constant for each error model of the spec, under
// the model's name, numbered once inside the range of its layer and never
// renumbered.
const (
	KindUnclassed1PathPoolsDoesnTPointTo                                  = LayerNoClass + 1  // 1001
	KindUnclassed1ShareAlreadyExists                                      = LayerNoClass + 2  // 1002
	KindUnclassed2PathPoolsIsAlreadyUsedIn                                = LayerNoClass + 3  // 1003
	KindUnclassed4ValueContainsInvalidCharactersOnlyA                     = LayerNoClass + 4  // 1004
	KindUnclassed4ValueHasInvalidLengthMaximumOf                          = LayerNoClass + 5  // 1005
	KindUnclassed4ValueHasInvalidLengthMinimumCharacter                   = LayerNoClass + 6  // 1006
	KindUnclassed6NotValidCharactersInARecycle                            = LayerNoClass + 7  // 1007
	KindKeyError_004371                                                   = LayerPython + 1   // 2001
	KindKeyError_40bcba                                                   = LayerPython + 2   // 2002
	KindOSError36ErrnoFileNameTooLongNfs                                  = LayerPython + 3   // 2003
	KindTypeErrorNotSupportedBetweenInstancesOfNonetype_a38cd7            = LayerPython + 4   // 2004
	KindTypeErrorNotSupportedBetweenInstancesOfNonetype_b1b8e2            = LayerPython + 5   // 2005
	KindUnicodeDecodeErrorUtfCodecCanTDecodeByte                          = LayerPython + 6   // 2006
	KindValueErrorDoesNotAppearToBeAn                                     = LayerPython + 7   // 2007
	KindValueErrorInvalidLiteralForIntWithBase                            = LayerPython + 8   // 2008
	KindValueErrorSizeSuffixGbIsIncorrect                                 = LayerPython + 9   // 2009
	KindValueErrorSizeUnitHasIncorrectValue                               = LayerPython + 10  // 2010
	KindPropertyError2CannotSetPropertyNosuchpropertyItNot                = LayerZfs + 1      // 3001
	KindZfsCmdError0CannotOpenDatasetDoesNotExist_96e556                  = LayerZfs + 2      // 3002
	KindZfsCmdError0CannotOpenDatasetDoesNotExist_a33d60                  = LayerZfs + 3      // 3003
	KindZfsCmdError1000CannotCreateInvalidNumericSuffix                   = LayerZfs + 4      // 3004
	KindZfsCmdError1000CannotCreateSnapshotBookmarkDelimiterIs            = LayerZfs + 5      // 3005
	KindZfsCmdError1000CannotCreateSnapshotEmptyComponentOr               = LayerZfs + 6      // 3006
	KindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_8fa47d      = LayerZfs + 7      // 3007
	KindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_cee0b6      = LayerZfs + 8      // 3008
	KindZfsCmdError1000CannotCreateSnapshotMultipleAndOr                  = LayerZfs + 9      // 3009
	KindZfsCmdError1000CannotCreateSnapshotTrailingSlashIn                = LayerZfs + 10     // 3010
	KindZfsCmdError1000CannotCreateSnapshotsThisPropertyCan               = LayerZfs + 11     // 3011
	KindZfsCmdError1000CannotCreateUseNoneToDisable                       = LayerZfs + 12     // 3012
	KindZfsCmdError1000CannotCreateVolumeSizeCannotBe                     = LayerZfs + 13     // 3013
	KindZfsCmdError1000CannotDestroySnapshotHasDependentClones            = LayerZfs + 14     // 3014
	KindZfsCmdError1000CannotDestroyVolumeHasChildren                     = LayerZfs + 15     // 3015
	KindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_214e41            = LayerZfs + 16     // 3016
	KindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_c9f5d6            = LayerZfs + 17     // 3017
	KindZfsCmdError1000CannotOpenInvalidCharacterInName_61dfda            = LayerZfs + 18     // 3018
	KindZfsCmdError1000CannotOpenInvalidCharacterInName_6d342b            = LayerZfs + 19     // 3019
	KindZfsCmdError1000CannotOpenMultipleAndOrDelimiters                  = LayerZfs + 20     // 3020
	KindZfsCmdError1000CannotOpenNameIsTooLong_37e45d                     = LayerZfs + 21     // 3021
	KindZfsCmdError1000CannotOpenNameIsTooLong_671b8f                     = LayerZfs + 22     // 3022
	KindZfsCmdError1000CannotOpenNameIsTooLong_a68ac4                     = LayerZfs + 23     // 3023
	KindZfsCmdError1000CannotOpenNameIsTooLong_d50fea                     = LayerZfs + 24     // 3024
	KindZfsCmdError1000WarningVolblocksizeIsLessThanTheVolumeSizeCannotBe = LayerZfs + 25     // 3025
	KindZfsCmdError100CannotCreateDatasetAlreadyExists                    = LayerZfs + 26     // 3026
	KindZfsCmdError201CannotCreateSnapshotNameIsToo_abcab7                = LayerZfs + 27     // 3027
	KindZfsCmdError201CannotCreateSnapshotNameIsToo_d5a47e                = LayerZfs + 28     // 3028
	KindZfsCmdError4CannotCreateOutOfSpace                                = LayerZfs + 29     // 3029
	KindZfsResourceError1ZfsResourceNotFoundInThis_3b952d                 = LayerZfs + 30     // 3030
	KindZfsResourceError1ZfsResourceNotFoundInThis_597b56                 = LayerZfs + 31     // 3031
	KindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec                 = LayerZfs + 32     // 3032
	KindZfsResourceError1ZfsResourceNotFoundInThis_a720ee                 = LayerZfs + 33     // 3033
	KindZfsResourceError4DatasetCanTContainCharacterIn                    = LayerZfs + 34     // 3034
	KindZfsResourceError4ZvolCanTContainCharacterIn                       = LayerZfs + 35     // 3035
	KindZfsResourceError5ResourceAlreadyExists                            = LayerZfs + 36     // 3036
	KindZfsResourceError5SnapshotAlreadyExists                            = LayerZfs + 37     // 3037
	KindZpoolError1GivenZpoolDoesnTExists                                 = LayerZfs + 38     // 3038
	KindAddressValueError                                                 = LayerProduct + 1  // 4001
	KindCalculationZvolSizeErrorVolumeSizeItsBlockSizeAnd                 = LayerProduct + 2  // 4002
	KindDatasetValidationError1DatasetNameTooLongDatasetName              = LayerProduct + 3  // 4003
	KindDatasetValidationError2DatasetNameIsInvalidAllowedCharacters      = LayerProduct + 4  // 4004
	KindItemConflictErrorLunIsAlreadyUsedIn                               = LayerProduct + 5  // 4005
	KindItemConflictErrorVolumeDevIsAlreadyUsed                           = LayerProduct + 6  // 4006
	KindItemNotFoundErrorPoolDoesNotExist                                 = LayerProduct + 7  // 4007
	KindItemNotFoundErrorPoolNotFound                                     = LayerProduct + 8  // 4008
	KindItemNotFoundErrorResourceNotFound                                 = LayerProduct + 9  // 4009
	KindItemNotFoundErrorTargetNotExists                                  = LayerProduct + 10 // 4010
	KindItemNotFoundErrorTargetPartitionNotExists                         = LayerProduct + 11 // 4011
	KindItemNotFoundErrorVolumeNameIsNotAttachedTo                        = LayerProduct + 12 // 4012
	KindItemNotFoundErrorVolumeNotFoundInPool                             = LayerProduct + 13 // 4013
	KindTargetNameConflictErrorTargetAlreadyExistsInScst                  = LayerProduct + 14 // 4014
	KindTargetNameConflictErrorTargetWithNameIqnOrgHurltest               = LayerProduct + 15 // 4015
	KindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4      = LayerProduct + 16 // 4016
	KindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_cfbcb7      = LayerProduct + 17 // 4017
	KindValidationError_7d2ccd                                            = LayerProduct + 18 // 4018
	KindZfsOeErrorInOrderToDeleteADataset                                 = LayerProduct + 19 // 4019
	KindZfsOeErrorInOrderToDeleteAZvol                                    = LayerProduct + 20 // 4020
	KindZfsOeErrorResourceIsNotADatasetType                               = LayerProduct + 21 // 4021
	KindZfsOeErrorResourceIsNotAZvolType                                  = LayerProduct + 22 // 4022
	KindZvolValidationError1ZvolNameTooLongZvolName                       = LayerProduct + 23 // 4023
	KindZvolValidationError2ZvolNameIsInvalidAllowedCharacters            = LayerProduct + 24 // 4024
	KindValidationError_6317af                                            = LayerSchema + 1   // 5001
	KindBadRequestAllowedValuesForBlocksizeAre                            = LayerRouting + 1  // 6001
	KindBadRequestAllowedValuesForCompressionAreOff                       = LayerRouting + 2  // 6002
	KindBadRequestAllowedValuesForCopiesAre                               = LayerRouting + 3  // 6003
	KindBadRequestAllowedValuesForDedupAreOff                             = LayerRouting + 4  // 6004
	KindBadRequestAllowedValuesForLogbiasAreLatency                       = LayerRouting + 5  // 6005
	KindBadRequestAllowedValuesForPrimarycacheAreAll                      = LayerRouting + 6  // 6006
	KindBadRequestAllowedValuesForSecondarycacheAreAll                    = LayerRouting + 7  // 6007
	KindBadRequestAllowedValuesForSyncAreAlways                           = LayerRouting + 8  // 6008
	KindBadRequestNewZvolSizeExceedsAvailableSpace                        = LayerRouting + 9  // 6009
	KindNotFound                                                          = LayerRouting + 10 // 6010
	KindNotFoundB_2e360f                                                  = LayerRouting + 11 // 6011
	KindNotFoundB_699120                                                  = LayerRouting + 12 // 6012
	KindNotFoundCloneNotExists_0d481d                                     = LayerRouting + 13 // 6013
	KindNotFoundShareNotExists                                            = LayerRouting + 14 // 6014
	KindNotFoundTb                                                        = LayerRouting + 15 // 6015
)

// What a reply of each kind looks like: its status, its class ("" where the
// spec lists the class as absent), its errno in its three cases and the
// pattern of its message (nil where the spec gives the kind no pattern).
var (
	// Path /Pools/{{pool}}/{{name}} doesn't point to NAS volume.
	kindUnclassed1PathPoolsDoesnTPointTo = errorKind{KindUnclassed1PathPoolsDoesnTPointTo, 400, "", errnoIs(1),
		regexp.MustCompile(`^Path /Pools/.*/.* doesn't point to NAS volume\.$`)}
	// Share {{body.name}} already exists.
	kindUnclassed1ShareAlreadyExists = errorKind{KindUnclassed1ShareAlreadyExists, 409, "", errnoIs(1),
		regexp.MustCompile(`^Share [\s\S]* already exists\.$`)}
	// Path /Pools/{{pool}}/{{name}} is already used in a different share.
	kindUnclassed2PathPoolsIsAlreadyUsedIn = errorKind{KindUnclassed2PathPoolsIsAlreadyUsedIn, 409, "", errnoIs(2),
		regexp.MustCompile(`^Path /Pools/.*/.* is already used in a different share\.$`)}
	// Value {{body.name}} contains invalid characters. Only a-z, A-Z, 0-9 . : $ _ and - are allowed.
	kindUnclassed4ValueContainsInvalidCharactersOnlyA = errorKind{KindUnclassed4ValueContainsInvalidCharactersOnlyA, 400, "", errnoIs(4),
		regexp.MustCompile(`^Value [\s\S]* contains invalid characters\. Only a-z, A-Z, 0-9 \. : \$ _ and - are allowed\.$`)}
	// Value has invalid length. Maximum of 255 characters are allowed.
	kindUnclassed4ValueHasInvalidLengthMaximumOf = errorKind{KindUnclassed4ValueHasInvalidLengthMaximumOf, 400, "", errnoIs(4),
		regexp.MustCompile(`^Value has invalid length\. Maximum of 255 characters are allowed\.$`)}
	// Value has invalid length. Minimum 1 character required.
	kindUnclassed4ValueHasInvalidLengthMinimumCharacter = errorKind{KindUnclassed4ValueHasInvalidLengthMinimumCharacter, 400, "", errnoIs(4),
		regexp.MustCompile(`^Value has invalid length\. Minimum 1 character required\.$`)}
	// Not valid characters in a recycle bin name. A name cannot contain any of the following charac...
	kindUnclassed6NotValidCharactersInARecycle = errorKind{KindUnclassed6NotValidCharactersInARecycle, 400, "", errnoIs(6),
		regexp.MustCompile("^Not valid characters in a recycle bin name\\. A name cannot contain any of the following characters & \\* / \\\\ : ; < > \\? ` \\| \"$")}
	// '{{query.search_by}}'
	kindKeyError_004371 = errorKind{KindKeyError_004371, 500, "builtins.KeyError", errnoMissing,
		regexp.MustCompile(`^'[\s\S]*'$`)}
	// '{{path.poolname}}/{{path.volumename}}@{{path.snapshotname}}'
	kindKeyError_40bcba = errorKind{KindKeyError_40bcba, 500, "builtins.KeyError", errnoMissing,
		regexp.MustCompile(`^'[\s\S]*/[\s\S]*@[\s\S]*'$`)}
	// [Errno 36] File name too long: '/NFS/{{body.name}}'
	kindOSError36ErrnoFileNameTooLongNfs = errorKind{KindOSError36ErrnoFileNameTooLongNfs, 500, "builtins.OSError", errnoIs(36),
		regexp.MustCompile(`^\[Errno 36\] File name too long: '/NFS/[\s\S]*'$`)}
	// '<' not supported between instances of 'NoneType' and 'NoneType'
	kindTypeErrorNotSupportedBetweenInstancesOfNonetype_a38cd7 = errorKind{KindTypeErrorNotSupportedBetweenInstancesOfNonetype_a38cd7, 500, "builtins.TypeError", errnoMissing,
		regexp.MustCompile(`^'<' not supported between instances of 'NoneType' and 'NoneType'$`)}
	// '<' not supported between instances of 'NoneType' and 'str'
	kindTypeErrorNotSupportedBetweenInstancesOfNonetype_b1b8e2 = errorKind{KindTypeErrorNotSupportedBetweenInstancesOfNonetype_b1b8e2, 500, "builtins.TypeError", errnoMissing,
		regexp.MustCompile(`^'<' not supported between instances of 'NoneType' and 'str'$`)}
	// 'utf-8' codec can't decode byte {{echo.byte}} in position {{echo.position}}: invalid continua...
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte = errorKind{KindUnicodeDecodeErrorUtfCodecCanTDecodeByte, 500, "builtins.UnicodeDecodeError", errnoMissing,
		regexp.MustCompile(`^'utf-8' codec can't decode byte [\s\S]* in position [\s\S]*: invalid continuation byte$`)}
	// '{{body.allow_ip}}' does not appear to be an IPv4 or IPv6 network
	kindValueErrorDoesNotAppearToBeAn = errorKind{KindValueErrorDoesNotAppearToBeAn, 500, "builtins.ValueError", errnoMissing,
		regexp.MustCompile(`^'[\s\S]*' does not appear to be an IPv4 or IPv6 network$`)}
	// invalid literal for int() with base 10: '{{query.page}}'
	kindValueErrorInvalidLiteralForIntWithBase = errorKind{KindValueErrorInvalidLiteralForIntWithBase, 500, "builtins.ValueError", errnoMissing,
		regexp.MustCompile(`^invalid literal for int\(\) with base 10: '[\s\S]*'$`)}
	// Size suffix GB is incorrect.
	kindValueErrorSizeSuffixGbIsIncorrect = errorKind{KindValueErrorSizeSuffixGbIsIncorrect, 500, "builtins.ValueError", errnoMissing,
		regexp.MustCompile(`^Size suffix GB is incorrect\.$`)}
	// Size unit {{body.blocksize}} has incorrect value.
	kindValueErrorSizeUnitHasIncorrectValue = errorKind{KindValueErrorSizeUnitHasIncorrectValue, 500, "builtins.ValueError", errnoMissing,
		regexp.MustCompile(`^Size unit [\s\S]* has incorrect value\.$`)}
	// Cannot set property nosuchproperty it not exists.
	kindPropertyError2CannotSetPropertyNosuchpropertyItNot = errorKind{KindPropertyError2CannotSetPropertyNosuchpropertyItNot, 500, "zfslib.zfsapi.properties.PropertyError", errnoIs(2),
		regexp.MustCompile(`^Cannot set property nosuchproperty it not exists\.$`)}
	// cannot open '{{path.poolname}}': dataset does not exist
	kindZfsCmdError0CannotOpenDatasetDoesNotExist_96e556 = errorKind{KindZfsCmdError0CannotOpenDatasetDoesNotExist_96e556, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(0),
		regexp.MustCompile(`^cannot open '[\s\S]*': dataset does not exist$`)}
	// cannot open '{{path.poolname}}/{{path.poolname}}@{{path.poolname}}': dataset does not exist
	kindZfsCmdError0CannotOpenDatasetDoesNotExist_a33d60 = errorKind{KindZfsCmdError0CannotOpenDatasetDoesNotExist_a33d60, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(0),
		regexp.MustCompile(`^cannot open '[\s\S]*/[\s\S]*@[\s\S]*': dataset does not exist$`)}
	// cannot create '{{path.poolname}}/{{body.name}}': invalid numeric suffix '{{echo.suffix}}'
	kindZfsCmdError1000CannotCreateInvalidNumericSuffix = errorKind{KindZfsCmdError1000CannotCreateInvalidNumericSuffix, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create '[\s\S]*/[\s\S]*': invalid numeric suffix '[\s\S]*'$`)}
	// cannot create snapshot '{{path.poolname}}/{{path.volumename}}@{{body.snapshot_name}}': bookma...
	kindZfsCmdError1000CannotCreateSnapshotBookmarkDelimiterIs = errorKind{KindZfsCmdError1000CannotCreateSnapshotBookmarkDelimiterIs, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create snapshot '[\s\S]*/[\s\S]*@[\s\S]*': bookmark delimiter '#' is not expected here$`)}
	// cannot create snapshot '{{path.poolname}}/{{path.volumename}}@': empty component or misplaced...
	kindZfsCmdError1000CannotCreateSnapshotEmptyComponentOr = errorKind{KindZfsCmdError1000CannotCreateSnapshotEmptyComponentOr, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create snapshot '[\s\S]*/[\s\S]*@': empty component or misplaced '@' or '#' delimiter in name$`)}
	// cannot create snapshot '{{path.poolname}}/{{path.volumename}}@{{body.snapshot_name}}': invali...
	kindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_8fa47d = errorKind{KindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_8fa47d, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create snapshot '[\s\S]*/[\s\S]*@[\s\S]*': invalid character % in name$`)}
	// cannot create snapshot '{{path.poolname}}/{{path.volumename}}@{{body.snapshot_name}}': invali...
	kindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_cee0b6 = errorKind{KindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_cee0b6, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create snapshot '[\s\S]*/[\s\S]*@[\s\S]*': invalid character '[\s\S]*' in name$`)}
	// cannot create snapshot '{{path.poolname}}/{{path.volumename}}@{{body.snapshot_name}}': multip...
	kindZfsCmdError1000CannotCreateSnapshotMultipleAndOr = errorKind{KindZfsCmdError1000CannotCreateSnapshotMultipleAndOr, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create snapshot '[\s\S]*/[\s\S]*@[\s\S]*': multiple '@' and/or '#' delimiters in name$`)}
	// cannot create snapshot '{{path.poolname}}/{{path.volumename}}@{{body.snapshot_name}}': traili...
	kindZfsCmdError1000CannotCreateSnapshotTrailingSlashIn = errorKind{KindZfsCmdError1000CannotCreateSnapshotTrailingSlashIn, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create snapshot '[\s\S]*/[\s\S]*@[\s\S]*': trailing slash in name$`)}
	// cannot create snapshots : this property can not be modified for snapshots
	kindZfsCmdError1000CannotCreateSnapshotsThisPropertyCan = errorKind{KindZfsCmdError1000CannotCreateSnapshotsThisPropertyCan, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create snapshots : this property can not be modified for snapshots$`)}
	// cannot create '{{path.poolname}}/{{body.name}}': use 'none' to disable quota/refquota
	kindZfsCmdError1000CannotCreateUseNoneToDisable = errorKind{KindZfsCmdError1000CannotCreateUseNoneToDisable, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create '[\s\S]*/[\s\S]*': use 'none' to disable quota/refquota$`)}
	// cannot create '{{path.poolname}}/{{body.name}}': volume size cannot be zero
	kindZfsCmdError1000CannotCreateVolumeSizeCannotBe = errorKind{KindZfsCmdError1000CannotCreateVolumeSizeCannotBe, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot create '[\s\S]*/[\s\S]*': volume size cannot be zero$`)}
	// cannot destroy '{{path.poolname}}/{{path.volumename}}@{{path.snapshotname}}': snapshot has de...
	kindZfsCmdError1000CannotDestroySnapshotHasDependentClones = errorKind{KindZfsCmdError1000CannotDestroySnapshotHasDependentClones, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot destroy '[\s\S]*/[\s\S]*@[\s\S]*': snapshot has dependent clones\nuse '-R' to destroy the following datasets:\n`)}
	// cannot destroy '{{path.poolname}}/{{path.volumename}}': volume has children\nuse '-r' to dest...
	kindZfsCmdError1000CannotDestroyVolumeHasChildren = errorKind{KindZfsCmdError1000CannotDestroyVolumeHasChildren, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot destroy '[\s\S]*/[\s\S]*': volume has children\nuse '-r' to destroy the following datasets:\n`)}
	// cannot open '{{path.poolname}}/{{path.volumename}}': bookmark delimiter '#' is not expected here
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_214e41 = errorKind{KindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_214e41, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot open '[\s\S]*/[\s\S]*': bookmark delimiter '#' is not expected here$`)}
	// cannot open '{{path.poolname}}/{{path.volumename}}@{{path.snapshotname}}': bookmark delimiter...
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_c9f5d6 = errorKind{KindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_c9f5d6, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot open '[\s\S]*/[\s\S]*@[\s\S]*': bookmark delimiter '#' is not expected here$`)}
	// cannot open '{{path.poolname}}/{{path.volumename}}@{{path.snapshotname}}': invalid character ...
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_61dfda = errorKind{KindZfsCmdError1000CannotOpenInvalidCharacterInName_61dfda, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot open '[\s\S]*/[\s\S]*@[\s\S]*': invalid character '[\s\S]*' in name$`)}
	// cannot open '{{path.poolname}}/{{path.volumename}}': invalid character '{{echo.char}}' in name
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_6d342b = errorKind{KindZfsCmdError1000CannotOpenInvalidCharacterInName_6d342b, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot open '[\s\S]*/[\s\S]*': invalid character '[\s\S]*' in name$`)}
	// cannot open '{{path.poolname}}/{{path.volumename}}@{{path.snapshotname}}': multiple '@' and/o...
	kindZfsCmdError1000CannotOpenMultipleAndOrDelimiters = errorKind{KindZfsCmdError1000CannotOpenMultipleAndOrDelimiters, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot open '[\s\S]*/[\s\S]*@[\s\S]*': multiple '@' and/or '#' delimiters in name$`)}
	// cannot open '{{path.poolname}}/{{name}}: name is too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_37e45d = errorKind{KindZfsCmdError1000CannotOpenNameIsTooLong_37e45d, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot open '[\s\S]*/.*: name is too long$`)}
	// cannot open '{{path.poolname}}/{{path.volumename}}@{{name}}: name is too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_671b8f = errorKind{KindZfsCmdError1000CannotOpenNameIsTooLong_671b8f, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot open '[\s\S]*/[\s\S]*@.*: name is too long$`)}
	// cannot open '{{path.poolname}}/{{path.volumename}}@{{path.snapshotname}}': name is too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_a68ac4 = errorKind{KindZfsCmdError1000CannotOpenNameIsTooLong_a68ac4, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot open '[\s\S]*/[\s\S]*@[\s\S]*': name is too long$`)}
	// cannot open '{{path.poolname}}/{{path.volumename}}': name is too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_d50fea = errorKind{KindZfsCmdError1000CannotOpenNameIsTooLong_d50fea, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^cannot open '[\s\S]*/[\s\S]*': name is too long$`)}
	// Warning: volblocksize ({{echo.blocksize}}) is less than the default minimum block size (16384...
	kindZfsCmdError1000WarningVolblocksizeIsLessThanTheVolumeSizeCannotBe = errorKind{KindZfsCmdError1000WarningVolblocksizeIsLessThanTheVolumeSizeCannotBe, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(1000),
		regexp.MustCompile(`^Warning: volblocksize \([\s\S]*\) is less than the default minimum block size \(16384\)\.\nTo reduce wasted space a volblocksize of 16384 is recommended\.\ncannot create '[\s\S]*/[\s\S]*': volume size cannot be zero$`)}
	// cannot create '{{path.poolname}}/{{body.name}}': dataset already exists
	kindZfsCmdError100CannotCreateDatasetAlreadyExists = errorKind{KindZfsCmdError100CannotCreateDatasetAlreadyExists, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(100),
		regexp.MustCompile(`^cannot create '[\s\S]*/[\s\S]*': dataset already exists$`)}
	// cannot create snapshot '{{path.poolname}}/{{path.volumename}}@{{name}}: name is too long
	kindZfsCmdError201CannotCreateSnapshotNameIsToo_abcab7 = errorKind{KindZfsCmdError201CannotCreateSnapshotNameIsToo_abcab7, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(201),
		regexp.MustCompile(`^cannot create snapshot '[\s\S]*/[\s\S]*@.*: name is too long$`)}
	// cannot create snapshot '{{path.poolname}}/{{path.volumename}}@{{body.snapshot_name}}': name i...
	kindZfsCmdError201CannotCreateSnapshotNameIsToo_d5a47e = errorKind{KindZfsCmdError201CannotCreateSnapshotNameIsToo_d5a47e, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(201),
		regexp.MustCompile(`^cannot create snapshot '[\s\S]*/[\s\S]*@[\s\S]*': name is too long$`)}
	// cannot create '{{path.poolname}}/{{body.name}}': out of space
	kindZfsCmdError4CannotCreateOutOfSpace = errorKind{KindZfsCmdError4CannotCreateOutOfSpace, 500, "zfslib.wrap.zfs.ZfsCmdError", errnoIs(4),
		regexp.MustCompile(`^cannot create '[\s\S]*/[\s\S]*': out of space$`)}
	// Zfs resource: {{path.poolname}}/{{path.poolname}} not found in this collection.
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d = errorKind{KindZfsResourceError1ZfsResourceNotFoundInThis_3b952d, 500, "zfslib.zfsapi.resources.ZfsResourceError", errnoIs(1),
		regexp.MustCompile(`^Zfs resource: [\s\S]*/[\s\S]* not found in this collection\.$`)}
	// Zfs resource: {{path.poolname}}@{{path.poolname}} not found in this collection.
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56 = errorKind{KindZfsResourceError1ZfsResourceNotFoundInThis_597b56, 500, "zfslib.zfsapi.resources.ZfsResourceError", errnoIs(1),
		regexp.MustCompile(`^Zfs resource: [\s\S]*@[\s\S]* not found in this collection\.$`)}
	// Zfs resource: {{path.poolname}}/{{path.volumename}}@{{path.snapshotname}} not found in this c...
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec = errorKind{KindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec, 500, "zfslib.zfsapi.resources.ZfsResourceError", errnoIs(1),
		regexp.MustCompile(`^Zfs resource: [\s\S]*/[\s\S]*@[\s\S]* not found in this collection\.$`)}
	// Zfs resource: {{path.poolname}}/{{name}} not found in this collection.
	kindZfsResourceError1ZfsResourceNotFoundInThis_a720ee = errorKind{KindZfsResourceError1ZfsResourceNotFoundInThis_a720ee, 500, "zfslib.zfsapi.resources.ZfsResourceError", errnoIs(1),
		regexp.MustCompile(`^Zfs resource: [\s\S]*/.* not found in this collection\.$`)}
	// Dataset can't contain '@' character in int's full name.
	kindZfsResourceError4DatasetCanTContainCharacterIn = errorKind{KindZfsResourceError4DatasetCanTContainCharacterIn, 500, "zfslib.zfsapi.resources.ZfsResourceError", errnoIs(4),
		regexp.MustCompile(`^Dataset can't contain '@' character in int's full name\.$`)}
	// Zvol can't contain '@' character in int's full name.
	kindZfsResourceError4ZvolCanTContainCharacterIn = errorKind{KindZfsResourceError4ZvolCanTContainCharacterIn, 500, "zfslib.zfsapi.resources.ZfsResourceError", errnoIs(4),
		regexp.MustCompile(`^Zvol can't contain '@' character in int's full name\.$`)}
	// Resource {{path.poolname}}/{{body.name}} already exists.
	kindZfsResourceError5ResourceAlreadyExists = errorKind{KindZfsResourceError5ResourceAlreadyExists, 500, "zfslib.zfsapi.resources.ZfsResourceError", errnoIs(5),
		regexp.MustCompile(`^Resource [\s\S]*/[\s\S]* already exists\.$`)}
	// Snapshot {{path.poolname}}/{{path.volumename}}@{{body.snapshot_name}} already exists.
	kindZfsResourceError5SnapshotAlreadyExists = errorKind{KindZfsResourceError5SnapshotAlreadyExists, 500, "zfslib.zfsapi.resources.ZfsResourceError", errnoIs(5),
		regexp.MustCompile(`^Snapshot [\s\S]*/[\s\S]*@[\s\S]* already exists\.$`)}
	// Given zpool '{{path.poolname}}' doesn't exists.
	kindZpoolError1GivenZpoolDoesnTExists = errorKind{KindZpoolError1GivenZpoolDoesnTExists, 500, "zfslib.zfsapi.zpool.ZpoolError", errnoIs(1),
		regexp.MustCompile(`^Given zpool '[\s\S]*' doesn't exists\.$`)}
	// {{body.nfs.allow_access_ip}}
	kindAddressValueError = errorKind{KindAddressValueError, 500, "opene.lib.ipaddr.AddressValueError", errnoMissing,
		regexp.MustCompile(`^[\s\S]*$`)}
	// volume size, its block size and amount of copies must be greater than 0
	kindCalculationZvolSizeErrorVolumeSizeItsBlockSizeAnd = errorKind{KindCalculationZvolSizeErrorVolumeSizeItsBlockSizeAnd, 500, "opene.tools.zreservation.CalculationZvolSizeError", errnoMissing,
		regexp.MustCompile(`^volume size, its block size and amount of copies must be greater than 0$`)}
	// Dataset name too long. Dataset name cannot exceed 132 characters.
	kindDatasetValidationError1DatasetNameTooLongDatasetName = errorKind{KindDatasetValidationError1DatasetNameTooLongDatasetName, 400, "opene.storage.zfs.exceptions.DatasetValidationError", errnoIs(1),
		regexp.MustCompile(`^Dataset name too long\. Dataset name cannot exceed 132 characters\.$`)}
	// Dataset name is invalid. Allowed characters in dataset name: a-z, A-Z, 0-9, _, -
	kindDatasetValidationError2DatasetNameIsInvalidAllowedCharacters = errorKind{KindDatasetValidationError2DatasetNameIsInvalidAllowedCharacters, 400, "opene.storage.zfs.exceptions.DatasetValidationError", errnoIs(2),
		regexp.MustCompile(`^Dataset name is invalid\. Allowed characters in dataset name: a-z, A-Z, 0-9, _, - $`)}
	// LUN 0 is already used in {{path.targetname}}.
	kindItemConflictErrorLunIsAlreadyUsedIn = errorKind{KindItemConflictErrorLunIsAlreadyUsedIn, 409, "opene.exceptions.ItemConflictError", errnoMissing,
		regexp.MustCompile(`^LUN 0 is already used in [\s\S]*\.$`)}
	// Volume /dev/{{path.poolname}}/{{body.name}} is already used.
	kindItemConflictErrorVolumeDevIsAlreadyUsed = errorKind{KindItemConflictErrorVolumeDevIsAlreadyUsed, 409, "opene.exceptions.ItemConflictError", errnoMissing,
		regexp.MustCompile(`^Volume /dev/[\s\S]*/[\s\S]* is already used\.$`)}
	// Pool {{path.poolname}} does not exist.
	kindItemNotFoundErrorPoolDoesNotExist = errorKind{KindItemNotFoundErrorPoolDoesNotExist, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Pool [\s\S]* does not exist\.$`)}
	// Pool {{path.poolname}} not found.
	kindItemNotFoundErrorPoolNotFound = errorKind{KindItemNotFoundErrorPoolNotFound, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Pool [\s\S]* not found\.$`)}
	// Resource {{path.poolname}}/{{path.volumename}} not found.
	kindItemNotFoundErrorResourceNotFound = errorKind{KindItemNotFoundErrorResourceNotFound, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Resource [\s\S]*/[\s\S]* not found\.$`)}
	// Target {{path.targetname}} not exists.
	kindItemNotFoundErrorTargetNotExists = errorKind{KindItemNotFoundErrorTargetNotExists, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Target [\s\S]* not exists\.$`)}
	// Target partition: {{path.poolname}} not exists.
	kindItemNotFoundErrorTargetPartitionNotExists = errorKind{KindItemNotFoundErrorTargetPartitionNotExists, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Target partition: [\s\S]* not exists\.$`)}
	// Volume name {{path.zvol}} is not attached to target {{path.targetname}}
	kindItemNotFoundErrorVolumeNameIsNotAttachedTo = errorKind{KindItemNotFoundErrorVolumeNameIsNotAttachedTo, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Volume name [\s\S]* is not attached to target [\s\S]*$`)}
	// Volume {{body.name}} not found in pool {{path.poolname}}.
	kindItemNotFoundErrorVolumeNotFoundInPool = errorKind{KindItemNotFoundErrorVolumeNotFoundInPool, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Volume [\s\S]* not found in pool [\s\S]*\.$`)}
	// Target already exists in SCST.
	kindTargetNameConflictErrorTargetAlreadyExistsInScst = errorKind{KindTargetNameConflictErrorTargetAlreadyExistsInScst, 409, "opene.san.target.base.iscsi.target.TargetNameConflictError", errnoMissing,
		regexp.MustCompile(`^Target already exists in SCST\.$`)}
	// Target with name iqn.2026-11.org.hurltest:{{name}} is already present on {{path.poolname}}.
	kindTargetNameConflictErrorTargetWithNameIqnOrgHurltest = errorKind{KindTargetNameConflictErrorTargetWithNameIqnOrgHurltest, 409, "opene.san.target.base.iscsi.target.TargetNameConflictError", errnoMissing,
		regexp.MustCompile(`^Target with name iqn\.2026-11\.org\.hurltest:.* is already present on [\s\S]*\.$`)}
	// Target with name {{body.name}} is already present on {{path.poolname}}.
	kindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4 = errorKind{KindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4, 409, "opene.san.target.base.iscsi.target.TargetNameConflictError", errnoMissing,
		regexp.MustCompile(`^Target with name [\s\S]* is already present on [\s\S]*\.$`)}
	// Target with name {{name}} is already present on {{path.poolname}}.
	kindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_cfbcb7 = errorKind{KindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_cfbcb7, 409, "opene.san.target.base.iscsi.target.TargetNameConflictError", errnoMissing,
		regexp.MustCompile(`^Target with name .* is already present on [\s\S]*\.$`)}
	// Target name is too long.
	kindValidationError_7d2ccd = errorKind{KindValidationError_7d2ccd, 400, "opene.exceptions.ValidationError", errnoMissing,
		nil}
	// In order to delete a dataset, you must delete all of its clones first.
	kindZfsOeErrorInOrderToDeleteADataset = errorKind{KindZfsOeErrorInOrderToDeleteADataset, 500, "opene.storage.zfs.zfs.ZfsOeError", errnoMissing,
		regexp.MustCompile(`^In order to delete a dataset, you must delete all of its clones first\.$`)}
	// In order to delete a zvol, you must delete all of its clones first.
	kindZfsOeErrorInOrderToDeleteAZvol = errorKind{KindZfsOeErrorInOrderToDeleteAZvol, 500, "opene.storage.zfs.zfs.ZfsOeError", errnoMissing,
		regexp.MustCompile(`^In order to delete a zvol, you must delete all of its clones first\.$`)}
	// Resource {{path.datasetname}} is not a dataset type.
	kindZfsOeErrorResourceIsNotADatasetType = errorKind{KindZfsOeErrorResourceIsNotADatasetType, 500, "opene.storage.zfs.zfs.ZfsOeError", errnoMissing,
		regexp.MustCompile(`^Resource [\s\S]* is not a dataset type\.$`)}
	// Resource {{path.volumename}} is not a zvol type.
	kindZfsOeErrorResourceIsNotAZvolType = errorKind{KindZfsOeErrorResourceIsNotAZvolType, 500, "opene.storage.zfs.zfs.ZfsOeError", errnoMissing,
		regexp.MustCompile(`^Resource [\s\S]* is not a zvol type\.$`)}
	// Zvol name too long. Zvol name cannot exceed 132 characters.
	kindZvolValidationError1ZvolNameTooLongZvolName = errorKind{KindZvolValidationError1ZvolNameTooLongZvolName, 400, "opene.storage.zfs.exceptions.ZvolValidationError", errnoIs(1),
		regexp.MustCompile(`^Zvol name too long\. Zvol name cannot exceed 132 characters\.$`)}
	// Zvol name is invalid. Allowed characters in zvol name: a-z, A-Z, 0-9, _, -
	kindZvolValidationError2ZvolNameIsInvalidAllowedCharacters = errorKind{KindZvolValidationError2ZvolNameIsInvalidAllowedCharacters, 400, "opene.storage.zfs.exceptions.ZvolValidationError", errnoIs(2),
		regexp.MustCompile(`^Zvol name is invalid\. Allowed characters in zvol name: a-z, A-Z, 0-9, _, - $`)}
	// 'name' is a required property
	kindValidationError_6317af = errorKind{KindValidationError_6317af, 400, "jsonschema.exceptions.ValidationError", errnoMissing,
		nil}
	// Allowed values for blocksize are: 4096, 8192, 16384, 32768, 65536, 131072, 262144, 524288, 10...
	kindBadRequestAllowedValuesForBlocksizeAre = errorKind{KindBadRequestAllowedValuesForBlocksizeAre, 400, "werkzeug.exceptions.BadRequest", errnoMissing,
		regexp.MustCompile(`^Allowed values for blocksize are: 4096, 8192, 16384, 32768, 65536, 131072, 262144, 524288, 1048576$`)}
	// Allowed values for compression are: off, on, lzjb, gzip, gzip-1, gzip-2, gzip-3, gzip-4, gzip...
	kindBadRequestAllowedValuesForCompressionAreOff = errorKind{KindBadRequestAllowedValuesForCompressionAreOff, 400, "werkzeug.exceptions.BadRequest", errnoMissing,
		regexp.MustCompile(`^Allowed values for compression are: off, on, lzjb, gzip, gzip-1, gzip-2, gzip-3, gzip-4, gzip-5, gzip-6, gzip-7, gzip-8, gzip-9, zle, lz4$`)}
	// Allowed values for copies are: 1, 2, 3
	kindBadRequestAllowedValuesForCopiesAre = errorKind{KindBadRequestAllowedValuesForCopiesAre, 400, "werkzeug.exceptions.BadRequest", errnoMissing,
		regexp.MustCompile(`^Allowed values for copies are: 1, 2, 3$`)}
	// Allowed values for dedup are: off, on, verify, sha256, sha256,verify
	kindBadRequestAllowedValuesForDedupAreOff = errorKind{KindBadRequestAllowedValuesForDedupAreOff, 400, "werkzeug.exceptions.BadRequest", errnoMissing,
		regexp.MustCompile(`^Allowed values for dedup are: off, on, verify, sha256, sha256,verify$`)}
	// Allowed values for logbias are: latency, throughput
	kindBadRequestAllowedValuesForLogbiasAreLatency = errorKind{KindBadRequestAllowedValuesForLogbiasAreLatency, 400, "werkzeug.exceptions.BadRequest", errnoMissing,
		regexp.MustCompile(`^Allowed values for logbias are: latency, throughput$`)}
	// Allowed values for primarycache are: all, none, metadata
	kindBadRequestAllowedValuesForPrimarycacheAreAll = errorKind{KindBadRequestAllowedValuesForPrimarycacheAreAll, 400, "werkzeug.exceptions.BadRequest", errnoMissing,
		regexp.MustCompile(`^Allowed values for primarycache are: all, none, metadata$`)}
	// Allowed values for secondarycache are: all, none, metadata
	kindBadRequestAllowedValuesForSecondarycacheAreAll = errorKind{KindBadRequestAllowedValuesForSecondarycacheAreAll, 400, "werkzeug.exceptions.BadRequest", errnoMissing,
		regexp.MustCompile(`^Allowed values for secondarycache are: all, none, metadata$`)}
	// Allowed values for sync are: always, standard, disabled
	kindBadRequestAllowedValuesForSyncAreAlways = errorKind{KindBadRequestAllowedValuesForSyncAreAlways, 400, "werkzeug.exceptions.BadRequest", errnoMissing,
		regexp.MustCompile(`^Allowed values for sync are: always, standard, disabled$`)}
	// New zvol size({{echo.size}}) exceeds available space on pool {{path.poolname}}({{echo.availab...
	kindBadRequestNewZvolSizeExceedsAvailableSpace = errorKind{KindBadRequestNewZvolSizeExceedsAvailableSpace, 400, "werkzeug.exceptions.BadRequest", errnoMissing,
		regexp.MustCompile(`^New zvol size\([\s\S]*\) exceeds available space on pool [\s\S]*\([\s\S]*\)\.$`)}
	// "{{path.clonename}}"
	kindNotFound = errorKind{KindNotFound, 404, "werkzeug.exceptions.NotFound", errnoMissing,
		regexp.MustCompile(`^"[\s\S]*"$`)}
	// '{{name}}\\b'
	kindNotFoundB_2e360f = errorKind{KindNotFoundB_2e360f, 404, "werkzeug.exceptions.NotFound", errnoMissing,
		regexp.MustCompile(`^'.*\\\\b'$`)}
	// '{{name}}\nb'
	kindNotFoundB_699120 = errorKind{KindNotFoundB_699120, 404, "werkzeug.exceptions.NotFound", errnoMissing,
		regexp.MustCompile(`^'.*\\nb'$`)}
	// Clone {{path.clonename}} not exists
	kindNotFoundCloneNotExists_0d481d = errorKind{KindNotFoundCloneNotExists_0d481d, 404, "werkzeug.exceptions.NotFound", errnoMissing,
		regexp.MustCompile(`^Clone [\s\S]* not exists$`)}
	// Share {{path.sharename}} not exists.
	kindNotFoundShareNotExists = errorKind{KindNotFoundShareNotExists, 404, "werkzeug.exceptions.NotFound", errnoMissing,
		regexp.MustCompile(`^Share [\s\S]* not exists\.$`)}
	// '{{name}}\tb'
	kindNotFoundTb = errorKind{KindNotFoundTb, 404, "werkzeug.exceptions.NotFound", errnoMissing,
		regexp.MustCompile(`^'.*\\tb'$`)}
)

// errorKinds lists every kind of error, in the order of its code.
var errorKinds = []errorKind{
	kindUnclassed1PathPoolsDoesnTPointTo,
	kindUnclassed1ShareAlreadyExists,
	kindUnclassed2PathPoolsIsAlreadyUsedIn,
	kindUnclassed4ValueContainsInvalidCharactersOnlyA,
	kindUnclassed4ValueHasInvalidLengthMaximumOf,
	kindUnclassed4ValueHasInvalidLengthMinimumCharacter,
	kindUnclassed6NotValidCharactersInARecycle,
	kindKeyError_004371,
	kindKeyError_40bcba,
	kindOSError36ErrnoFileNameTooLongNfs,
	kindTypeErrorNotSupportedBetweenInstancesOfNonetype_a38cd7,
	kindTypeErrorNotSupportedBetweenInstancesOfNonetype_b1b8e2,
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte,
	kindValueErrorDoesNotAppearToBeAn,
	kindValueErrorInvalidLiteralForIntWithBase,
	kindValueErrorSizeSuffixGbIsIncorrect,
	kindValueErrorSizeUnitHasIncorrectValue,
	kindPropertyError2CannotSetPropertyNosuchpropertyItNot,
	kindZfsCmdError0CannotOpenDatasetDoesNotExist_96e556,
	kindZfsCmdError0CannotOpenDatasetDoesNotExist_a33d60,
	kindZfsCmdError1000CannotCreateInvalidNumericSuffix,
	kindZfsCmdError1000CannotCreateSnapshotBookmarkDelimiterIs,
	kindZfsCmdError1000CannotCreateSnapshotEmptyComponentOr,
	kindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_8fa47d,
	kindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_cee0b6,
	kindZfsCmdError1000CannotCreateSnapshotMultipleAndOr,
	kindZfsCmdError1000CannotCreateSnapshotTrailingSlashIn,
	kindZfsCmdError1000CannotCreateSnapshotsThisPropertyCan,
	kindZfsCmdError1000CannotCreateUseNoneToDisable,
	kindZfsCmdError1000CannotCreateVolumeSizeCannotBe,
	kindZfsCmdError1000CannotDestroySnapshotHasDependentClones,
	kindZfsCmdError1000CannotDestroyVolumeHasChildren,
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_214e41,
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_c9f5d6,
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_61dfda,
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_6d342b,
	kindZfsCmdError1000CannotOpenMultipleAndOrDelimiters,
	kindZfsCmdError1000CannotOpenNameIsTooLong_37e45d,
	kindZfsCmdError1000CannotOpenNameIsTooLong_671b8f,
	kindZfsCmdError1000CannotOpenNameIsTooLong_a68ac4,
	kindZfsCmdError1000CannotOpenNameIsTooLong_d50fea,
	kindZfsCmdError1000WarningVolblocksizeIsLessThanTheVolumeSizeCannotBe,
	kindZfsCmdError100CannotCreateDatasetAlreadyExists,
	kindZfsCmdError201CannotCreateSnapshotNameIsToo_abcab7,
	kindZfsCmdError201CannotCreateSnapshotNameIsToo_d5a47e,
	kindZfsCmdError4CannotCreateOutOfSpace,
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d,
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56,
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec,
	kindZfsResourceError1ZfsResourceNotFoundInThis_a720ee,
	kindZfsResourceError4DatasetCanTContainCharacterIn,
	kindZfsResourceError4ZvolCanTContainCharacterIn,
	kindZfsResourceError5ResourceAlreadyExists,
	kindZfsResourceError5SnapshotAlreadyExists,
	kindZpoolError1GivenZpoolDoesnTExists,
	kindAddressValueError,
	kindCalculationZvolSizeErrorVolumeSizeItsBlockSizeAnd,
	kindDatasetValidationError1DatasetNameTooLongDatasetName,
	kindDatasetValidationError2DatasetNameIsInvalidAllowedCharacters,
	kindItemConflictErrorLunIsAlreadyUsedIn,
	kindItemConflictErrorVolumeDevIsAlreadyUsed,
	kindItemNotFoundErrorPoolDoesNotExist,
	kindItemNotFoundErrorPoolNotFound,
	kindItemNotFoundErrorResourceNotFound,
	kindItemNotFoundErrorTargetNotExists,
	kindItemNotFoundErrorTargetPartitionNotExists,
	kindItemNotFoundErrorVolumeNameIsNotAttachedTo,
	kindItemNotFoundErrorVolumeNotFoundInPool,
	kindTargetNameConflictErrorTargetAlreadyExistsInScst,
	kindTargetNameConflictErrorTargetWithNameIqnOrgHurltest,
	kindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4,
	kindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_cfbcb7,
	kindValidationError_7d2ccd,
	kindZfsOeErrorInOrderToDeleteADataset,
	kindZfsOeErrorInOrderToDeleteAZvol,
	kindZfsOeErrorResourceIsNotADatasetType,
	kindZfsOeErrorResourceIsNotAZvolType,
	kindZvolValidationError1ZvolNameTooLongZvolName,
	kindZvolValidationError2ZvolNameIsInvalidAllowedCharacters,
	kindValidationError_6317af,
	kindBadRequestAllowedValuesForBlocksizeAre,
	kindBadRequestAllowedValuesForCompressionAreOff,
	kindBadRequestAllowedValuesForCopiesAre,
	kindBadRequestAllowedValuesForDedupAreOff,
	kindBadRequestAllowedValuesForLogbiasAreLatency,
	kindBadRequestAllowedValuesForPrimarycacheAreAll,
	kindBadRequestAllowedValuesForSecondarycacheAreAll,
	kindBadRequestAllowedValuesForSyncAreAlways,
	kindBadRequestNewZvolSizeExceedsAvailableSpace,
	kindNotFound,
	kindNotFoundB_2e360f,
	kindNotFoundB_699120,
	kindNotFoundCloneNotExists_0d481d,
	kindNotFoundShareNotExists,
	kindNotFoundTb,
}

// Entries not in the spec: errors known from elsewhere than the jdss1 v4
// scenarios files. Such an entry has no kind: the second code of its error
// is that of the layer of the reply's class alone. They are listed to the
// jdss-rest-api-spec project as to be added.
//
// TODO(T13): none of them has a recorded reply on jdss1 v4, so none has a
// fixture and none is tested; that needs a test server that answers with
// such errors (kub-testing ISSUES.md, T13).
var (
	// `Resource <p>/<n> already exists.` with errno 1: a pattern of the code before API v4 that no
	// recording supports, most likely from an older release. No recording gives its status or its
	// class, so the entry takes any.
	notInSpecResourceExistsErrno1 = errorEntry{status: anyStatus, class: anyClass, errno: errnoIs(1),
		message: regexp.MustCompile(`Resource .*\/.* already exists\.`),
		code:    RestErrorResourceExists}
	// The same message with errno 100.
	notInSpecResourceExistsErrno100 = errorEntry{status: anyStatus, class: anyClass, errno: errnoIs(100),
		message: regexp.MustCompile(`Resource .*\/.* already exists\.`),
		code:    RestErrorResourceExists}
	// `Zfs resource: <p>/<n> not found in this collection` without an errno: a pattern of the code
	// before API v4 that no recording supports. Every recorded reply of this kind carries errno 1.
	notInSpecZfsResourceNotFoundNoErrno = errorEntry{status: anyStatus, class: anyClass, errno: errnoMissing,
		message: regexp.MustCompile(`Zfs resource: (.+\/.+) not found in this collection`),
		code:    RestErrorResourceDNE}
	// Out of space on the zvol create: recorded on jdss1 v3 and on jdss2, not on jdss1 v4, where a
	// zvol create on a pool too small is answered 400, `New zvol size(...) exceeds available space`.
	notInSpecZvolCreateOutOfSpace = errorEntry{status: 500, class: "zfslib.wrap.zfs.ZfsCmdError", errno: errnoIs(4),
		message: regexp.MustCompile(`^cannot create '[\s\S]*/[\s\S]*': out of space$`),
		code:    RestErrorOutOfSpace}
	// Out of space on the NAS snapshot create: recorded on jdss1 v3 and on jdss2, in no jdss1 v4
	// scenarios file yet.
	notInSpecNASSnapshotCreateOutOfSpace = errorEntry{status: 500, class: "zfslib.wrap.zfs.ZfsCmdError", errno: errnoIs(1000),
		message: regexp.MustCompile(`^cannot create snapshots : out of space$`),
		code:    RestErrorOutOfSpace}
)

// GetPool, GET pools/<poolname>.
var getPoolErrors = []errorEntry{
	kindItemNotFoundErrorPoolDoesNotExist.means(RestErrorResourceDNEPool), // the pool is missing
}

// GetVolume, GET pools/<poolname>/volumes/<volumename>.
var getVolumeErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_214e41.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_6d342b.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError1000CannotOpenNameIsTooLong_37e45d
	kindZfsCmdError1000CannotOpenNameIsTooLong_d50fea.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_37e45d.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d.means(RestErrorResourceDNE),
	kindZfsResourceError4ZvolCanTContainCharacterIn.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// CreateVolume, POST pools/<poolname>/volumes.
var createVolumeErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindValueErrorSizeSuffixGbIsIncorrect.means(RestErrorFailureUnknown),
	kindValueErrorSizeUnitHasIncorrectValue.means(RestErrorFailureUnknown),
	kindPropertyError2CannotSetPropertyNosuchpropertyItNot.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateVolumeSizeCannotBe.means(RestErrorFailureUnknown),
	kindZfsCmdError1000WarningVolblocksizeIsLessThanTheVolumeSizeCannotBe.means(RestErrorFailureUnknown),
	kindZfsCmdError100CannotCreateDatasetAlreadyExists.means(RestErrorResourceExists),
	kindZfsResourceError5ResourceAlreadyExists.means(RestErrorResourceExists),
	kindZpoolError1GivenZpoolDoesnTExists.means(RestErrorResourceDNEPool), // the pool is missing
	kindCalculationZvolSizeErrorVolumeSizeItsBlockSizeAnd.means(RestErrorFailureUnknown),
	kindZvolValidationError1ZvolNameTooLongZvolName.means(RestErrorArgumentIncorrect), // a name too long
	kindZvolValidationError2ZvolNameIsInvalidAllowedCharacters.means(RestErrorFailureUnknown),
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	kindBadRequestAllowedValuesForBlocksizeAre.means(RestErrorFailureUnknown),
	kindBadRequestAllowedValuesForCompressionAreOff.means(RestErrorFailureUnknown),
	kindBadRequestAllowedValuesForCopiesAre.means(RestErrorFailureUnknown),
	kindBadRequestAllowedValuesForDedupAreOff.means(RestErrorFailureUnknown),
	kindBadRequestAllowedValuesForLogbiasAreLatency.means(RestErrorFailureUnknown),
	kindBadRequestAllowedValuesForPrimarycacheAreAll.means(RestErrorFailureUnknown),
	kindBadRequestAllowedValuesForSecondarycacheAreAll.means(RestErrorFailureUnknown),
	kindBadRequestAllowedValuesForSyncAreAlways.means(RestErrorFailureUnknown),
	kindBadRequestNewZvolSizeExceedsAvailableSpace.means(RestErrorOutOfSpace),
	// not in the spec
	notInSpecZvolCreateOutOfSpace,
	notInSpecResourceExistsErrno1,
	notInSpecResourceExistsErrno100,
}

// DeleteVolume, DELETE pools/<poolname>/volumes/<volumename>.
var deleteVolumeErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotDestroyVolumeHasChildren.means(RestErrorResourceBusyVolumeHasSnapshots),
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_214e41.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_6d342b.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError1000CannotOpenNameIsTooLong_37e45d
	kindZfsCmdError1000CannotOpenNameIsTooLong_d50fea.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_37e45d.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d.means(RestErrorResourceDNE),
	kindZfsResourceError4ZvolCanTContainCharacterIn.means(RestErrorFailureUnknown),
	kindZfsOeErrorInOrderToDeleteAZvol.means(RestErrorResourceBusy),
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// GetVolumesEntries, GET pools/<poolname>/volumes.
var getVolumesEntriesErrors = []errorEntry{
	kindKeyError_004371.means(RestErrorFailureUnknown),
	kindValueErrorInvalidLiteralForIntWithBase.means(RestErrorFailureUnknown),
	kindZpoolError1GivenZpoolDoesnTExists.means(RestErrorResourceDNEPool), // the pool is missing
}

// GetVolumeSnapshot, GET pools/<poolname>/volumes/<volumename>/snapshots/<snapshotname>.
var getVolumeSnapshotErrors = []errorEntry{
	kindKeyError_40bcba.means(RestErrorFailureUnknown),
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError0CannotOpenDatasetDoesNotExist_a33d60.means(RestErrorResourceDNE),
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_61dfda.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenMultipleAndOrDelimiters.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError1000CannotOpenNameIsTooLong_671b8f
	kindZfsCmdError1000CannotOpenNameIsTooLong_a68ac4.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_671b8f.means(RestErrorArgumentIncorrect), // a name too long
}

// CreateSnapshot, POST pools/<poolname>/volumes/<volumename>/snapshots.
var createSnapshotErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotBookmarkDelimiterIs.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotEmptyComponentOr.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_8fa47d.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_cee0b6.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotMultipleAndOr.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotTrailingSlashIn.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotsThisPropertyCan.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError201CannotCreateSnapshotNameIsToo_abcab7
	kindZfsCmdError201CannotCreateSnapshotNameIsToo_d5a47e.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError201CannotCreateSnapshotNameIsToo_abcab7.means(RestErrorArgumentIncorrect), // a name too long
	// TODO: a missing pool gives this error too, and cannot be told from the
	// missing parent here: taken as the parent missing (ADR 0002, question 48)
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d.means(RestErrorResourceDNE),
	kindZfsResourceError5SnapshotAlreadyExists.means(RestErrorResourceExists), // a repeated snapshot create
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecResourceExistsErrno1,
	notInSpecResourceExistsErrno100,
	notInSpecZfsResourceNotFoundNoErrno,
}

// DeleteSnapshot, DELETE pools/<poolname>/volumes/<volumename>/snapshots/<snapshotname>.
var deleteSnapshotErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotDestroySnapshotHasDependentClones.means(RestErrorResourceBusySnapshotHasClones),
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_c9f5d6.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_61dfda.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenMultipleAndOrDelimiters.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError1000CannotOpenNameIsTooLong_671b8f
	kindZfsCmdError1000CannotOpenNameIsTooLong_a68ac4.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_671b8f.means(RestErrorArgumentIncorrect), // a name too long
	// before these, whose patterns match this message too:
	//   ZfsResourceError1ZfsResourceNotFoundInThis_597b56
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec.means(RestErrorResourceDNE),
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56.means(RestErrorResourceDNEPool), // the pool is missing
	kindZfsOeErrorResourceIsNotAZvolType.means(RestErrorFailureUnknown),
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// GetSnapshotsEntries, GET pools/<poolname>/volumes/snapshots.
var getSnapshotsEntriesErrors = []errorEntry{
	kindValueErrorInvalidLiteralForIntWithBase.means(RestErrorFailureUnknown),
	kindItemNotFoundErrorPoolNotFound.means(RestErrorResourceDNEPool), // the pool is missing
}

// GetVolumeSnapshotsEntries, GET pools/<poolname>/volumes/<volumename>/snapshots.
var getVolumeSnapshotsEntriesErrors = []errorEntry{
	kindValueErrorInvalidLiteralForIntWithBase.means(RestErrorFailureUnknown),
	kindItemNotFoundErrorPoolNotFound.means(RestErrorResourceDNEPool), // the pool is missing
	kindItemNotFoundErrorResourceNotFound.means(RestErrorResourceDNE), // the volume whose snapshots are listed is missing
}

// GetVolumeSnapshotClones, GET pools/<poolname>/volumes/<volumename>/snapshots/<snapshotname>/clones.
var getVolumeSnapshotClonesErrors = []errorEntry{
	// before these, whose patterns match this message too:
	//   ZfsResourceError1ZfsResourceNotFoundInThis_597b56
	//   ZfsResourceError1ZfsResourceNotFoundInThis_a720ee
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec.means(RestErrorResourceDNE),
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56.means(RestErrorResourceDNEPool), // the pool is missing
	kindZfsResourceError1ZfsResourceNotFoundInThis_a720ee.means(RestErrorResourceDNE),
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// CreateClone, POST pools/<poolname>/volumes/<volumename>/clone.
var createCloneErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError100CannotCreateDatasetAlreadyExists.means(RestErrorResourceExists),
	// before these, whose patterns match this message too:
	//   ZfsResourceError1ZfsResourceNotFoundInThis_3b952d
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec.means(RestErrorResourceDNE),
	// TODO: a missing pool gives this error too, and cannot be told from the
	// missing parent here: taken as the parent missing (ADR 0002, question 48)
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d.means(RestErrorResourceDNE),
	kindZvolValidationError1ZvolNameTooLongZvolName.means(RestErrorArgumentIncorrect), // a name too long
	kindZvolValidationError2ZvolNameIsInvalidAllowedCharacters.means(RestErrorFailureUnknown),
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecResourceExistsErrno1,
	notInSpecResourceExistsErrno100,
	notInSpecZfsResourceNotFoundNoErrno,
}

// DeleteClone, DELETE pools/<poolname>/volumes/<volumename>/snapshots/<snapshotname>/clones/<clonename>.
var deleteCloneErrors = []errorEntry{
	kindZfsCmdError1000CannotDestroyVolumeHasChildren.means(RestErrorResourceBusyVolumeHasSnapshots),
	// before these, whose patterns match this message too:
	//   ZfsResourceError1ZfsResourceNotFoundInThis_3b952d
	//   ZfsResourceError1ZfsResourceNotFoundInThis_597b56
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec.means(RestErrorResourceDNE),
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d.means(RestErrorResourceDNE),
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56.means(RestErrorResourceDNEPool), // the pool is missing
	kindZfsOeErrorInOrderToDeleteAZvol.means(RestErrorResourceBusy),
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// GetTarget, GET pools/<poolname>/san/iscsi/targets/<targetname>.
var getTargetErrors = []errorEntry{
	// before these, whose patterns match this message too:
	//   ItemNotFoundErrorTargetNotExists
	kindItemNotFoundErrorTargetPartitionNotExists.means(RestErrorResourceDNEPool), // the pool is missing
	kindItemNotFoundErrorTargetNotExists.means(RestErrorResourceDNETarget),        // the target is missing
}

// CreateTarget, POST pools/<poolname>/san/iscsi/targets.
var createTargetErrors = []errorEntry{
	kindValueErrorDoesNotAppearToBeAn.means(RestErrorFailureUnknown),
	kindItemNotFoundErrorTargetPartitionNotExists.means(RestErrorResourceDNEPool), // the pool is missing
	kindTargetNameConflictErrorTargetAlreadyExistsInScst.means(RestErrorFailureUnknown),
	kindValidationError_7d2ccd.means(RestErrorFailureUnknown),
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   TargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4
	//   TargetNameConflictErrorTargetWithNameIsAlreadyPresent_cfbcb7
	// the spec gives no way to tell this kind from the one below, by status,
	// class, errno and pattern; a reply of that kind gets this second code:
	//   TargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4
	kindTargetNameConflictErrorTargetWithNameIqnOrgHurltest.means(RestErrorResourceExists),
	// before these, whose patterns match this message too:
	//   TargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4
	kindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_cfbcb7.means(RestErrorResourceExists),
	kindTargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4.means(RestErrorResourceExists),
}

// DeleteTarget, DELETE pools/<poolname>/san/iscsi/targets/<targetname>.
var deleteTargetErrors = []errorEntry{
	// before these, whose patterns match this message too:
	//   ItemNotFoundErrorTargetNotExists
	kindItemNotFoundErrorTargetPartitionNotExists.means(RestErrorResourceDNEPool), // the pool is missing
	kindItemNotFoundErrorTargetNotExists.means(RestErrorResourceDNETarget),        // the target is missing
}

// AttachVolumeToTarget, POST pools/<poolname>/san/iscsi/targets/<targetname>/luns.
var attachVolumeToTargetErrors = []errorEntry{
	kindItemConflictErrorLunIsAlreadyUsedIn.means(RestErrorResourceExists),
	kindItemConflictErrorVolumeDevIsAlreadyUsed.means(RestErrorResourceBusy),
	// before these, whose patterns match this message too:
	//   ItemNotFoundErrorTargetNotExists
	kindItemNotFoundErrorTargetPartitionNotExists.means(RestErrorResourceDNEPool), // the pool is missing
	kindItemNotFoundErrorTargetNotExists.means(RestErrorResourceDNETarget),        // the target is missing
	kindItemNotFoundErrorVolumeNotFoundInPool.means(RestErrorResourceDNEVolume),
	kindValidationError_7d2ccd.means(RestErrorFailureUnknown),
	kindValidationError_6317af.means(RestErrorFailureUnknown),
}

// DettachVolumeFromTarget, DELETE pools/<poolname>/san/iscsi/targets/<targetname>/luns/<zvol>.
var dettachVolumeFromTargetErrors = []errorEntry{
	// before these, whose patterns match this message too:
	//   ItemNotFoundErrorTargetNotExists
	kindItemNotFoundErrorTargetPartitionNotExists.means(RestErrorResourceDNEPool), // the pool is missing
	kindItemNotFoundErrorTargetNotExists.means(RestErrorResourceDNETarget),        // the target is missing
	kindItemNotFoundErrorVolumeNameIsNotAttachedTo.means(RestErrorResourceDNELun), // the LUN is gone; the target may still be there
}

// GetNASVolume, GET pools/<poolname>/nas-volumes/<datasetname>.
var getNASVolumeErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_214e41.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_6d342b.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError1000CannotOpenNameIsTooLong_37e45d
	kindZfsCmdError1000CannotOpenNameIsTooLong_d50fea.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_37e45d.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d.means(RestErrorResourceDNE),
	kindZfsResourceError4DatasetCanTContainCharacterIn.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// CreateNASVolume, POST pools/<poolname>/nas-volumes.
var createNASVolumeErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError0CannotOpenDatasetDoesNotExist_96e556.means(RestErrorResourceDNEPool), // the pool is missing
	kindZfsCmdError1000CannotCreateInvalidNumericSuffix.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateUseNoneToDisable.means(RestErrorFailureUnknown),
	kindZfsCmdError100CannotCreateDatasetAlreadyExists.means(RestErrorResourceExists),
	kindZfsCmdError4CannotCreateOutOfSpace.means(RestErrorOutOfSpace), // the pool is full
	kindZfsResourceError5ResourceAlreadyExists.means(RestErrorResourceExists),
	kindDatasetValidationError1DatasetNameTooLongDatasetName.means(RestErrorArgumentIncorrect), // a name too long
	kindDatasetValidationError2DatasetNameIsInvalidAllowedCharacters.means(RestErrorFailureUnknown),
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecResourceExistsErrno1,
	notInSpecResourceExistsErrno100,
}

// DeleteNASVolume, DELETE pools/<poolname>/nas-volumes/<datasetname>.
var deleteNASVolumeErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_214e41.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_6d342b.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError1000CannotOpenNameIsTooLong_37e45d
	kindZfsCmdError1000CannotOpenNameIsTooLong_d50fea.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_37e45d.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d.means(RestErrorResourceDNE),
	kindZfsResourceError4DatasetCanTContainCharacterIn.means(RestErrorFailureUnknown),
	kindZfsOeErrorInOrderToDeleteADataset.means(RestErrorResourceBusy), // a NAS volume with clones, as a zvol with clones
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// GetNASVolumesEntries, GET pools/<poolname>/nas-volumes.
var getNASVolumesEntriesErrors = []errorEntry{
	kindKeyError_004371.means(RestErrorFailureUnknown),
	kindTypeErrorNotSupportedBetweenInstancesOfNonetype_a38cd7.means(RestErrorFailureUnknown),
	kindTypeErrorNotSupportedBetweenInstancesOfNonetype_b1b8e2.means(RestErrorFailureUnknown),
	kindValueErrorInvalidLiteralForIntWithBase.means(RestErrorFailureUnknown),
	kindZpoolError1GivenZpoolDoesnTExists.means(RestErrorResourceDNEPool), // the pool is missing
}

// GetNASVolumeSnapshotsEntries, GET pools/<poolname>/nas-volumes/<datasetname>/snapshots.
var getNASVolumeSnapshotsEntriesErrors = []errorEntry{
	kindValueErrorInvalidLiteralForIntWithBase.means(RestErrorFailureUnknown),
	kindItemNotFoundErrorPoolNotFound.means(RestErrorResourceDNEPool), // the pool is missing
	kindItemNotFoundErrorResourceNotFound.means(RestErrorResourceDNE), // the volume whose snapshots are listed is missing
}

// GetNASSnapshotsEntries, GET pools/<poolname>/nas-volumes/snapshots.
var getNASSnapshotsEntriesErrors = []errorEntry{
	kindValueErrorInvalidLiteralForIntWithBase.means(RestErrorFailureUnknown),
	kindItemNotFoundErrorPoolNotFound.means(RestErrorResourceDNEPool), // the pool is missing
}

// CreateNASSnapshot, POST pools/<poolname>/nas-volumes/<datasetname>/snapshots.
var createNASSnapshotErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotBookmarkDelimiterIs.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotEmptyComponentOr.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_8fa47d.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_cee0b6.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotMultipleAndOr.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotCreateSnapshotTrailingSlashIn.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError201CannotCreateSnapshotNameIsToo_abcab7
	kindZfsCmdError201CannotCreateSnapshotNameIsToo_d5a47e.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError201CannotCreateSnapshotNameIsToo_abcab7.means(RestErrorArgumentIncorrect), // a name too long
	// TODO: a missing pool gives this error too, and cannot be told from the
	// missing parent here: taken as the parent missing (ADR 0002, question 48)
	kindZfsResourceError1ZfsResourceNotFoundInThis_3b952d.means(RestErrorResourceDNE),
	kindZfsResourceError5SnapshotAlreadyExists.means(RestErrorResourceExists), // a repeated snapshot create
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecNASSnapshotCreateOutOfSpace,
	notInSpecResourceExistsErrno1,
	notInSpecResourceExistsErrno100,
	notInSpecZfsResourceNotFoundNoErrno,
}

// DeleteNASSnapshot, DELETE pools/<poolname>/nas-volumes/<datasetname>/snapshots/<snapshotname>.
var deleteNASSnapshotErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_c9f5d6.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_61dfda.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenMultipleAndOrDelimiters.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError1000CannotOpenNameIsTooLong_671b8f
	kindZfsCmdError1000CannotOpenNameIsTooLong_a68ac4.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_671b8f.means(RestErrorArgumentIncorrect), // a name too long
	// before these, whose patterns match this message too:
	//   ZfsResourceError1ZfsResourceNotFoundInThis_597b56
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec.means(RestErrorResourceDNE),
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56.means(RestErrorResourceDNEPool), // the pool is missing
	kindZfsOeErrorResourceIsNotADatasetType.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// GetNASVolumeSnapshot, GET pools/<poolname>/nas-volumes/<datasetname>/snapshots/<snapshotname>.
var getNASVolumeSnapshotErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_c9f5d6.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenInvalidCharacterInName_61dfda.means(RestErrorFailureUnknown),
	kindZfsCmdError1000CannotOpenMultipleAndOrDelimiters.means(RestErrorFailureUnknown),
	// before these, whose patterns match this message too:
	//   ZfsCmdError1000CannotOpenNameIsTooLong_671b8f
	kindZfsCmdError1000CannotOpenNameIsTooLong_a68ac4.means(RestErrorArgumentIncorrect), // a name too long
	kindZfsCmdError1000CannotOpenNameIsTooLong_671b8f.means(RestErrorArgumentIncorrect), // a name too long
	// before these, whose patterns match this message too:
	//   ZfsResourceError1ZfsResourceNotFoundInThis_597b56
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec.means(RestErrorResourceDNE),
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56.means(RestErrorResourceDNEPool), // the pool is missing
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// GetNASVolumeSnapshotClones, GET pools/<poolname>/nas-volumes/<datasetname>/snapshots/<snapshotname>/clones.
var getNASVolumeSnapshotClonesErrors = []errorEntry{
	// before these, whose patterns match this message too:
	//   ZfsResourceError1ZfsResourceNotFoundInThis_597b56
	//   ZfsResourceError1ZfsResourceNotFoundInThis_a720ee
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec.means(RestErrorResourceDNE),
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56.means(RestErrorResourceDNEPool), // the pool is missing
	kindZfsResourceError1ZfsResourceNotFoundInThis_a720ee.means(RestErrorResourceDNE),
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// CreateNASClone, POST pools/<poolname>/nas-volumes/<datasetname>/snapshots/<snapshotname>/clones.
var createNASCloneErrors = []errorEntry{
	kindUnicodeDecodeErrorUtfCodecCanTDecodeByte.means(RestErrorFailureUnknown),
	kindZfsCmdError100CannotCreateDatasetAlreadyExists.means(RestErrorResourceExists),
	// before these, whose patterns match this message too:
	//   ZfsResourceError1ZfsResourceNotFoundInThis_597b56
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec.means(RestErrorResourceDNE),
	// TODO: a missing pool gives this error too, and cannot be told from the
	// missing parent here: taken as the parent missing (ADR 0002, question 48)
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56.means(RestErrorResourceDNE),
	kindDatasetValidationError1DatasetNameTooLongDatasetName.means(RestErrorArgumentIncorrect), // a name too long
	kindDatasetValidationError2DatasetNameIsInvalidAllowedCharacters.means(RestErrorFailureUnknown),
	kindValidationError_6317af.means(RestErrorFailureUnknown),
	// not in the spec
	notInSpecResourceExistsErrno1,
	notInSpecResourceExistsErrno100,
	notInSpecZfsResourceNotFoundNoErrno,
}

// DeleteNASClone, DELETE pools/<poolname>/nas-volumes/<datasetname>/snapshots/<snapshotname>/clones/<clonename>.
var deleteNASCloneErrors = []errorEntry{
	// before these, whose patterns match this message too:
	//   ZfsResourceError1ZfsResourceNotFoundInThis_597b56
	kindZfsResourceError1ZfsResourceNotFoundInThis_8f51ec.means(RestErrorResourceDNE),
	kindZfsResourceError1ZfsResourceNotFoundInThis_597b56.means(RestErrorResourceDNEPool), // the pool is missing
	kindNotFound.means(RestErrorResourceDNE),                                              // the clone is missing
	kindNotFoundB_2e360f.means(RestErrorResourceDNE),                                      // the clone is missing
	kindNotFoundB_699120.means(RestErrorResourceDNE),                                      // the clone is missing
	kindNotFoundCloneNotExists_0d481d.means(RestErrorResourceDNE),                         // the clone is missing
	kindNotFoundTb.means(RestErrorResourceDNE),                                            // the clone is missing
	// not in the spec
	notInSpecZfsResourceNotFoundNoErrno,
}

// CreateShare, POST shares.
var createShareErrors = []errorEntry{
	kindUnclassed1PathPoolsDoesnTPointTo.means(RestErrorFailureUnknown),
	kindUnclassed1ShareAlreadyExists.means(RestErrorResourceExists), // a repeated share create; the caller reads the share and compares its path
	kindUnclassed2PathPoolsIsAlreadyUsedIn.means(RestErrorFailureUnknown),
	kindUnclassed4ValueContainsInvalidCharactersOnlyA.means(RestErrorFailureUnknown),
	kindUnclassed4ValueHasInvalidLengthMaximumOf.means(RestErrorFailureUnknown),
	kindUnclassed4ValueHasInvalidLengthMinimumCharacter.means(RestErrorFailureUnknown),
	kindUnclassed6NotValidCharactersInARecycle.means(RestErrorFailureUnknown),
	kindOSError36ErrnoFileNameTooLongNfs.means(RestErrorArgumentIncorrect), // a name too long
	kindAddressValueError.means(RestErrorFailureUnknown),
	kindValidationError_7d2ccd.means(RestErrorFailureUnknown),
	kindValidationError_6317af.means(RestErrorFailureUnknown),
}

// GetShare, GET shares/<sharename>.
var getShareErrors = []errorEntry{
	kindNotFoundShareNotExists.means(RestErrorResourceDNEShare),
}

// errorCatalogue gives the catalogue of each call by the name of the call.
var errorCatalogue = map[string][]errorEntry{
	"GetPool":                      getPoolErrors,
	"GetVolume":                    getVolumeErrors,
	"CreateVolume":                 createVolumeErrors,
	"DeleteVolume":                 deleteVolumeErrors,
	"GetVolumesEntries":            getVolumesEntriesErrors,
	"GetVolumeSnapshot":            getVolumeSnapshotErrors,
	"CreateSnapshot":               createSnapshotErrors,
	"DeleteSnapshot":               deleteSnapshotErrors,
	"GetSnapshotsEntries":          getSnapshotsEntriesErrors,
	"GetVolumeSnapshotsEntries":    getVolumeSnapshotsEntriesErrors,
	"GetVolumeSnapshotClones":      getVolumeSnapshotClonesErrors,
	"CreateClone":                  createCloneErrors,
	"DeleteClone":                  deleteCloneErrors,
	"GetTarget":                    getTargetErrors,
	"CreateTarget":                 createTargetErrors,
	"DeleteTarget":                 deleteTargetErrors,
	"AttachVolumeToTarget":         attachVolumeToTargetErrors,
	"DettachVolumeFromTarget":      dettachVolumeFromTargetErrors,
	"GetNASVolume":                 getNASVolumeErrors,
	"CreateNASVolume":              createNASVolumeErrors,
	"DeleteNASVolume":              deleteNASVolumeErrors,
	"GetNASVolumesEntries":         getNASVolumesEntriesErrors,
	"GetNASVolumeSnapshotsEntries": getNASVolumeSnapshotsEntriesErrors,
	"GetNASSnapshotsEntries":       getNASSnapshotsEntriesErrors,
	"CreateNASSnapshot":            createNASSnapshotErrors,
	"DeleteNASSnapshot":            deleteNASSnapshotErrors,
	"GetNASVolumeSnapshot":         getNASVolumeSnapshotErrors,
	"GetNASVolumeSnapshotClones":   getNASVolumeSnapshotClonesErrors,
	"CreateNASClone":               createNASCloneErrors,
	"DeleteNASClone":               deleteNASCloneErrors,
	"CreateShare":                  createShareErrors,
	"GetShare":                     getShareErrors,
}
