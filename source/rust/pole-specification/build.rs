use std::io::{Error, Result};
use std::path::PathBuf;

fn main() -> Result<()> {
    let proto_dir: PathBuf = "proto".into(); // rust proto file dir
    let mut proto_list = Vec::new();
    for entry in proto_dir.read_dir()? {
        let path = entry?.path();
        if path.is_file()
            && path
                .extension()
                .is_some_and(|extension| extension == "proto")
        {
            proto_list.push(path);
        }
    }
    // 固定输入顺序，避免文件系统遍历顺序改变合并 package 的生成源码。
    proto_list.sort();

    tonic_prost_build::configure()
        .build_server(true)
        .out_dir("src")
        .compile_protos(&proto_list, &["proto/".into()])
        .map_err(Error::other)?;
    Ok(())
}
