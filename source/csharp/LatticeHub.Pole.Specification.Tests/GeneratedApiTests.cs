using Google.Protobuf;
using Pole.Sidecar.V1;
using Xunit;

namespace LatticeHub.Pole.Specification.Tests;

public sealed class GeneratedApiTests
{
    [Fact]
    public void BootstrapMessagesRoundTrip()
    {
        var hello = new ClientHello
        {
            SdkLanguage = "csharp",
            SdkVersion = "0.1.0"
        };
        hello.SupportedProtocols.AddRange([
            Protocol.Http,
            Protocol.Grpc,
            Protocol.Dubbo,
            Protocol.Thrift
        ]);

        var decoded = ClientHello.Parser.ParseFrom(hello.ToByteArray());

        Assert.Equal("csharp", decoded.SdkLanguage);
        Assert.Equal(4, decoded.SupportedProtocols.Count);
    }

    [Fact]
    public void BootstrapGrpcServiceIsGenerated()
    {
        Assert.Equal(
            "pole.sidecar.v1.SidecarSessionService",
            SidecarSessionService.Descriptor.FullName);
        Assert.True(SidecarSessionService.Descriptor.Methods[0].IsServerStreaming);
        Assert.Equal("OpenSession", SidecarSessionService.Descriptor.Methods[0].Name);
    }

    [Fact]
    public void EveryAuthoritativeProtoHasGeneratedReflectionMetadata()
    {
        var reflectionTypes = typeof(ClientHello).Assembly
            .GetTypes()
            .Where(type => type.IsAbstract && type.IsSealed && type.Name.EndsWith("Reflection"))
            .ToArray();

        Assert.Equal(26, reflectionTypes.Length);
    }
}
