data "netapp-ontap_storage_snapshot_policy_data_source" "storage_snapshot_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "ansible2"
}
