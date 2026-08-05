#include <google/protobuf/descriptor.h>

#include "bootstrap.grpc.pb.h"
#include "bootstrap.pb.h"

int main() {
    const auto* message = pole::sidecar::v1::ClientHello::descriptor();
    if (message == nullptr || message->full_name() != "pole.sidecar.v1.ClientHello") {
        return 1;
    }

    const auto* service = google::protobuf::DescriptorPool::generated_pool()
                              ->FindServiceByName("pole.sidecar.v1.SidecarSessionService");
    if (service == nullptr || service->method_count() != 2) {
        return 2;
    }
    const auto* control_session = service->FindMethodByName("OpenControlSession");
    if (control_session == nullptr || !control_session->client_streaming() ||
        !control_session->server_streaming()) {
        return 3;
    }

    using Stub = pole::sidecar::v1::SidecarSessionService::Stub;
    static_cast<void>(sizeof(Stub));
    return 0;
}
