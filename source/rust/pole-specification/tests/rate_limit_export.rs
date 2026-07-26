use pole_specification::polaris::metric::v2::rate_limit_grpcv2_client::RateLimitGrpcv2Client;

fn assert_public_type<T>() {}

#[test]
fn rate_limit_grpc_client_is_exported_from_the_public_module_path() {
    assert_public_type::<RateLimitGrpcv2Client<tonic::transport::Channel>>();
}
