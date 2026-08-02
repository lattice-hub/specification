use pole_specification::polaris::metric::v2::{
    rate_limit_grpc_client::RateLimitGrpcClient, QuotaAccounting, QuotaConsumption,
    QuotaReserveRequest, QuotaSettleRequest, QuotaSettlement, QuotaUpdateRequest, RateLimitCmd,
};

fn assert_public_type<T>() {}

#[test]
fn rate_limit_grpc_client_is_exported_from_the_public_module_path() {
    assert_public_type::<RateLimitGrpcClient<tonic::transport::Channel>>();
}

#[test]
fn quota_lease_protocol_is_exported_from_the_public_module_path() {
    assert_eq!(RateLimitCmd::Reserve as i32, 1);
    assert_eq!(RateLimitCmd::Update as i32, 3);
    assert_eq!(RateLimitCmd::Settle as i32, 4);
    assert_eq!(QuotaAccounting::Consumable as i32, 0);
    assert_eq!(QuotaAccounting::Occupancy as i32, 1);
    assert_public_type::<QuotaReserveRequest>();
    assert_public_type::<QuotaConsumption>();
    assert_public_type::<QuotaUpdateRequest>();
    assert_public_type::<QuotaSettleRequest>();
    assert_public_type::<QuotaSettlement>();
}
