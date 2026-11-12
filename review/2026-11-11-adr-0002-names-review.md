# ADR 0002: the names against the naming rule of 2026-11-11

Date: 2026-11-11, night. The user expanded the naming rule: a name exists so that a human reader can keep track of what is referenced, where letter-and-number IDs are hard to remember and get confused; it is given with care, catches the essence of the thing precisely enough that no later item can be confused with it, and is no longer than that takes; for a REST call, the API version, the path with its placeholders as words, the method, then what is wrong (bad `bug_plugin_rest_nas_volume_keys_never_match`, good `bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec`). The rule is now in `docs/adr/README.md` and `kub-testing/issues/README.md`. This report reads every name of the record (77 with a heading, 13 in code font) against it, with the entry's content as the measure of "the essence", and proposes a name where the current one fails. Report only: a name never changes without the user's decision, and the registry rule is that a changed name is a new row with the old one marked superseded.

## Verdict

38 of 90 names fail the rule, 52 meet it. The failures are of three kinds: the response-type and error-message bugs name a symptom in the plugin where the endpoint and the field are what pin them (the user's own example is one of these); a dozen logic bugs name a symptom without the function it lives in; and five implementation items carry names that any later item of the same kind would collide with ("crash fixes", "remove dead code", "tls and transport"). Three entries each cover several endpoints or several findings; for those a split gives the path form, and a single name on the common essence is the alternative. The crashes, gaps, rules, security findings and the names without a heading meet the rule as they are.

The proposed names follow the user's structure without exception, `pools_pool` prefix included, so four of them are long; the paths of the clones and the LUNs are five and six segments deep and nothing shorter pins them. Where a shorter name would do the user may trim, but a consistent shape across the names was given priority.

The cost of a rename is the number of places the name is written: the record (headings and links), the kub-testing files (the registry, `ISSUES.md`, the test cases) and the reviews, which are not rewritten. The columns below give those counts.

## Names that fail the rule, with a proposed name

| Current name | Verdict | Proposed name | Note | record | kub-testing | reviews |
|---|---|---|---|---|---|---|
| `bug_plugin_rest_volume_encryption_type` | rename | `bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch` | `encryption` is an object; the struct says string | 33 | 1 | 0 |
| `bug_plugin_rest_pool_iostats_type` | rename | `bug_rest_v4_pools_pool_get_iostats_type_mismatch` | numbers sent, strings expected; pool, vdev and disk | 31 | 3 | 0 |
| `bug_plugin_rest_disk_led_type` | rename | `bug_rest_v4_pools_pool_get_disk_led_type_mismatch` | `led` is an object; the struct says string | 25 | 1 | 0 |
| `bug_plugin_rest_snapshot_creation_type` | rename | `bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch` | also the per-volume list; a read field | 31 | 1 | 1 |
| `bug_plugin_rest_clone_is_clone_type` | rename | `bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch` | long: the clones path is five segments deep | 25 | 1 | 0 |
| `bug_plugin_rest_nas_volume_keys_never_match` | rename | `bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec` | the user's example | 19 | 2 | 1 |
| `bug_plugin_rest_snapshot_list_volume_key` | rename | `bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead` | the key the struct expects does not exist | 23 | 1 | 0 |
| `bug_plugin_rest_snapshot_kind_wrong_key` | rename | `bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key` | a decoder defect, so the area stays `plugin_rest` | 15 | 1 | 0 |
| `bug_plugin_rest_zvol_snapshot_name_missing` | rename | `bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key` |  | 9 | 1 | 4 |
| `bug_jdss_rest_share_not_exists_style` | rename | `bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class` | appliance behaviour, jdss2 only | 5 | 1 | 0 |
| `bug_plugin_rest_snapshot_exists_unrecognised` | split or rename | `bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised` | two endpoints: the NAS one is `…_nas_volumes_dataset_snapshots_post_…`; one entry per endpoint, or one name on the common essence | 19 | 1 | 1 |
| `bug_plugin_rest_detach_not_found_variants` | rename | `bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone` | long: the full path is the only thing that pins it | 23 | 1 | 0 |
| `bug_plugin_rest_nas_delete_with_clones_unrecognised` | rename | `bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised` |  | 17 | 1 | 0 |
| `bug_plugin_rest_share_exists_unrecognised` | rename | `bug_rest_v4_shares_post_409_already_exists_unrecognised` |  | 19 | 1 | 0 |
| `bug_plugin_rest_nas_clone_not_exists_unrecognised` | rename | `bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised` | long: the deepest path of the API | 17 | 1 | 0 |
| `bug_plugin_rest_out_of_space_messages_unrecognised` | split or rename | `bug_rest_v4_create_calls_out_of_space_messages_unrecognised` | three messages over the zvol, NAS and snapshot creates; one entry per call would give the path form | 17 | 1 | 0 |
| `bug_plugin_rest_missing_pool_or_parent_errors` | split or rename | `bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised` | three findings in one entry: the missing-pool 500, the targets 404, the name-too-long 400; a split names each | 19 | 1 | 2 |
| `bug_plugin_rest_target_missing_test_pattern` | rename | `bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name` | the defect is the pattern; its effect is on `GET`/`DELETE pools/<pool>/san/iscsi/targets/<target>` 404 | 11 | 2 | 0 |
| `bug_plugin_rest_volume_snapshots_stub` | rename | `bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub` |  | 21 | 1 | 1 |
| `bug_plugin_rest_nas_snapshots_wrong_path` | rename | `bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots` |  | 19 | 1 | 0 |
| `bug_plugin_rest_nas_snapshot_delete_method` | rename | `bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete` |  | 29 | 1 | 0 |
| `bug_plugin_rest_decode_before_status_check` | rename | `bug_plugin_rest_get_calls_decode_body_before_204_check` |  | 25 | 1 | 1 |
| `bug_plugin_rest_create_target_202_polling` | rename | `bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay` |  | 15 | 2 | 0 |
| `bug_plugin_rest_nas_snapshot_delete_decode_before_204` | rename | `bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check` |  | 11 | 1 | 0 |
| `bug_plugin_rest_parse_error_swallowed` | rename | `bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors` |  | 9 | 1 | 0 |
| `bug_plugin_rest_status_ignored_in_error_mapping` | rename | `bug_plugin_rest_get_error_maps_body_only_ignores_http_status` |  | 19 | 1 | 0 |
| `bug_plugin_rest_list_paging_without_per_page` | rename | `bug_plugin_rest_list_calls_send_page_without_per_page` |  | 13 | 1 | 0 |
| `bug_plugin_rest_volume_create_sparse_unset` | rename | `bug_plugin_rest_create_volume_never_sets_sparse` |  | 13 | 1 | 1 |
| `bug_plugin_driver_delete_cleanup_only_on_busy` | rename | `bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children` |  | 9 | 1 | 0 |
| `bug_plugin_driver_cleanup_type_check_before_error` | rename | `bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error` |  | 13 | 1 | 1 |
| `bug_plugin_rest_tries_and_addrs_unused` | rename | `bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover` |  | 7 | 1 | 0 |
| `bug_plugin_rest_no_request_deadline` | rename | `bug_plugin_rest_request_built_without_context_csi_deadline_not_applied` | mild: the cause is what pins it | 7 | 1 | 0 |
| `gap_jdss_rest_connected_clients` | rename | `gap_jdss_rest_detach_and_delete_with_connected_clients` | mild: says what was never recorded | 13 | 1 | 0 |
| `impl_plugin_rest_crash_fixes` | rename | `impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes` | any later crash fix in `pkg/rest` would collide with the current name | 21 | 1 | 5 |
| `impl_plugin_rest_fixtures_and_unit_tests` | rename | `impl_plugin_rest_recorded_reply_fixtures` | the tests are the implementing agent's since 2026-11-11 | 39 | 1 | 3 |
| `impl_plugin_rest_remove_dead_code` | rename | `impl_plugin_rest_remove_dead_code_before_v4_move` | dead code will be removed again some day | 3 | 1 | 0 |
| `impl_plugin_rest_request_side_check` | rename | `impl_plugin_rest_request_body_keys_checked_against_schema` | mild | 3 | 1 | 1 |
| `impl_plugin_rest_tls_and_transport` | rename | `impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile` | mild: "transport" could be anything later | 41 | 1 | 0 |

## Names that meet the rule

| Name | record | kub-testing | reviews |
|---|---|---|---|
| `bug_plugin_rest_reply_text_as_format_string` | 5 | 1 | 0 |
| `bug_plugin_rest_unknown_error_empty_message` | 5 | 1 | 0 |
| `bug_plugin_rest_every_error_logged_at_warn` | 5 | 1 | 0 |
| `crash_plugin_rest_nil_response_on_transport_error` | 15 | 1 | 0 |
| `crash_plugin_rest_basic_auth_before_request_check` | 9 | 1 | 0 |
| `crash_plugin_rest_error_from_nil_body` | 13 | 1 | 0 |
| `crash_plugin_rest_setup_endpoint_ignores_error` | 9 | 1 | 0 |
| `crash_plugin_rest_endpoint_string_empty_addrs` | 7 | 1 | 2 |
| `gap_jdss_rest_empty_body_on_get_delete` | 11 | 2 | 1 |
| `gap_jdss_rest_list_paging_without_per_page` | 9 | 1 | 0 |
| `gap_jdss_rest_volume_create_without_sparse` | 9 | 1 | 0 |
| `gap_jdss_rest_name_length_limit` | 5 | 1 | 0 |
| `gap_jdss_rest_out_of_space_messages` | 5 | 1 | 0 |
| `gap_jdss_rest_delete_missing_zvol_clone` | 3 | 1 | 0 |
| `gap_jdss_rest_clones_of_missing_parent` | 5 | 1 | 0 |
| `gap_jdss_rest_missing_pool_replies` | 5 | 1 | 1 |
| `gap_plugin_rest_appliance_unreachable` | 9 | 1 | 1 |
| `gap_plugin_rest_unexpected_reply_bodies` | 7 | 1 | 1 |
| `gap_plugin_rest_failover_during_call` | 7 | 1 | 1 |
| `rule_rest_call_exists_on_v4` | 9 | 1 | 1 |
| `rule_rest_request_body_valid` | 13 | 1 | 0 |
| `rule_rest_success_statuses_accepted` | 9 | 1 | 0 |
| `rule_rest_error_statuses_handled` | 13 | 1 | 0 |
| `rule_rest_reply_types_decode` | 13 | 1 | 0 |
| `rule_rest_error_bodies_handled` | 7 | 1 | 1 |
| `rule_rest_docs_vs_appliance_listed` | 11 | 1 | 0 |
| `impl_plugin_rest_error_catalogue` | 35 | 1 | 3 |
| `impl_plugin_rest_escape_path_segments` | 15 | 1 | 2 |
| `impl_plugin_rest_generated_reply_types` | 41 | 1 | 1 |
| `impl_plugin_rest_no_format_string_from_reply` | 17 | 1 | 0 |
| `impl_plugin_rest_no_request_logging` | 19 | 1 | 5 |
| `impl_plugin_rest_package_documentation` | 3 | 1 | 0 |
| `impl_plugin_rest_reply_size_limit` | 15 | 1 | 2 |
| `impl_plugin_rest_return_parse_errors` | 7 | 1 | 3 |
| `impl_plugin_driver_validate_names_and_ids` | 43 | 2 | 2 |
| `impl_plugin_no_secrets_in_logs` | 21 | 1 | 0 |
| `impl_lab_rest_audit_tool` | 19 | 1 | 0 |
| `security_plugin_rest_tls_verification_off` | 5 | 1 | 0 |
| `security_plugin_rest_unescaped_path_segments` | 9 | 1 | 0 |
| `security_plugin_rest_request_dump_every_call` | 4 | 1 | 4 |
| `security_plugin_secrets_in_logs` | 6 | 1 | 4 |
| `bug_plugin_driver_nas_quota_null_size_zero` | 3 | 1 | 1 |
| `bug_plugin_node_iscsi_login_failure_ignored` | 4 | 1 | 1 |
| `bug_plugin_node_unstage_no_logout` | 4 | 1 | 1 |
| `bug_plugin_rest_delete_share_nil_descriptor` | 2 | 1 | 1 |
| `bug_plugin_rest_nas_snapshot_create_copy_of_volume_create` | 2 | 1 | 1 |
| `crash_plugin_ctrl_create_snapshot_empty_name` | 4 | 1 | 1 |
| `crash_plugin_ctrl_create_volume_on_rest_error` | 2 | 1 | 1 |
| `crash_plugin_ctrl_get_capacity_no_logger` | 4 | 1 | 1 |
| `crash_plugin_driver_id_index_without_length_check` | 4 | 1 | 1 |
| `crash_plugin_driver_snapshot_name_empty` | 4 | 1 | 1 |
| `crash_plugin_driver_volume_desc_empty_value` | 4 | 1 | 1 |

## If the user decides to rename

- The registry rule applies: each new name is a new row in its `kub-testing/issues/` file, the old row is marked superseded, and the old ID stays an alias; the registry README's history line is the precedent (`work` became `impl` on 2026-11-11).
- In the record, the `####` heading and every link change together (the README's Navigation rule); `tools/check_names.py` then shows zero unregistered names and `tools/check_record.py` zero linked names without a heading.
- `kub-testing/ISSUES.md` and the test cases that cite a name change in the same commit; the reviews keep the old names, which the registry maps.
- The three "split or rename" entries need a decision first: one entry per endpoint, or one name on the common essence.
- The record's prose that explains a name by its old wording (for example "keys never match") is re-read where the name changes, so that the text and the name say the same thing.

## Outcome

Decided on 2026-11-12: long names are accepted, and the 38 names were renamed as proposed, the three "split or rename" entries under their single proposed name. Applied the same day: the record (688 mentions: headings, links and the ID sources table), the registry (40 rows in five files, each now the current row with the old row under it marked superseded and listing the reviews that still use the old name; two names sat under two old IDs), the examples of both READMEs, the ADR README's exception to the never-changed rule, the notes file, and `check_names.py`, which now reads the registry rows and flags a superseded name written in the record (section H, exit 1). `kub-testing/ISSUES.md` and the test cases held none of the 38 names; the reviews listed above keep the old names, which the superseded rows map. The registry's meaning of F27 was reworded to match the record (the plugin recognises none of the three replies; the old wording said "recognised, but the callers misread it"). The entries were re-read under their new names: all say what the name says, except that the entry now named `bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised` also covers the NAS snapshot create, whose path the name does not carry; that is backlog item 53. Both checkers pass.
