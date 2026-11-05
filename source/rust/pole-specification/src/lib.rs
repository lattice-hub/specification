#[rustfmt::skip]
pub mod v1;

pub mod pole {
    /// 限流数据面；v2 是该既有协议的版本，与配置中心 v1 独立。
    pub mod metric {
        pub mod v2 {
            include!("pole.metric.v2.rs");
        }
    }
    /// 配置中心统一定义、发布及服务端解析结果协议。
    pub mod config {
        pub mod v1 {
            include!("pole.config.v1.rs");
        }
    }
    pub mod sidecar {
        pub mod v1 {
            include!("pole.sidecar.v1.rs");
        }
    }
}

#[test]
fn it_works() {
    println!("ok");
}
