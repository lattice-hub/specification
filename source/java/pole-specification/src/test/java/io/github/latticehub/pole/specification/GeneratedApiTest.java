package io.github.latticehub.pole.specification;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;

import com.google.protobuf.DescriptorProtos.FileDescriptorSet;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Set;
import java.util.TreeSet;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Stream;
import org.junit.jupiter.api.Test;

class GeneratedApiTest {
    private static final Pattern JAVA_PACKAGE =
            Pattern.compile("option\\s+java_package\\s*=\\s*\"([^\"]+)\"");
    private static final Pattern SERVICE =
            Pattern.compile("(?m)^\\s*service\\s+([A-Za-z][A-Za-z0-9_]*)\\s*\\{");

    @Test
    void compilesEveryAuthoritativeProtoAndGrpcService() throws Exception {
        Path apiDirectory = Path.of("../../../api").toAbsolutePath().normalize();
        Set<String> authoritativeProtoNames = protoNames(apiDirectory);
        Set<String> flattenedProtoNames = protoNames(Path.of("target/generated-proto"));
        assertFalse(authoritativeProtoNames.isEmpty());
        try (Stream<Path> paths = Files.walk(apiDirectory)) {
            long authoritativeProtoCount =
                    paths.filter(path -> path.toString().endsWith(".proto")).count();
            assertEquals(authoritativeProtoCount, authoritativeProtoNames.size(),
                    "duplicate proto basename");
        }
        assertEquals(authoritativeProtoNames, flattenedProtoNames);

        Path descriptorPath = Path.of(
                "target/generated-resources/protobuf/pole-specification.desc");
        FileDescriptorSet descriptorSet =
                FileDescriptorSet.parseFrom(Files.readAllBytes(descriptorPath));
        Set<String> compiledProtoNames = new TreeSet<>();
        descriptorSet.getFileList().forEach(file -> compiledProtoNames.add(file.getName()));
        assertEquals(authoritativeProtoNames, compiledProtoNames);

        try (Stream<Path> paths = Files.walk(apiDirectory)) {
            for (Path proto : paths.filter(path -> path.toString().endsWith(".proto")).toList()) {
                assertGrpcServicesLoad(proto);
            }
        }
    }

    private static Set<String> protoNames(Path directory) throws IOException {
        Set<String> names = new TreeSet<>();
        try (Stream<Path> paths = Files.walk(directory)) {
            paths.filter(path -> path.toString().endsWith(".proto"))
                    .forEach(path -> names.add(path.getFileName().toString()));
        }
        return names;
    }

    private static void assertGrpcServicesLoad(Path proto) throws Exception {
        String source = Files.readString(proto);
        Matcher packageMatcher = JAVA_PACKAGE.matcher(source);
        if (!packageMatcher.find()) {
            throw new AssertionError("missing java_package: " + proto);
        }
        String javaPackage = packageMatcher.group(1);
        Matcher serviceMatcher = SERVICE.matcher(source);
        while (serviceMatcher.find()) {
            Class.forName(javaPackage + "." + serviceMatcher.group(1) + "Grpc");
        }
    }
}
