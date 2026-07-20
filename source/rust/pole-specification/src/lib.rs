pub mod v1;

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
