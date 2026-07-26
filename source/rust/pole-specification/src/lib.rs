pub mod v1;

/// 限流数据面 gRPC 协议。
///
/// 限流规则通过 `v1::RateLimit::cluster` 选择目标集群，SDK 使用这里的双向流协议
/// 初始化配额、上报使用量并接收服务器下发的可用配额。协议定义已经由
/// specification 生成，crate 仅在此公开稳定的 Rust 模块路径。
pub mod polaris {
    pub mod metric {
        pub mod v2 {
            include!("polaris.metric.v2.rs");
        }
    }
}

#[test]
fn it_works() {
    println!("ok");
}
