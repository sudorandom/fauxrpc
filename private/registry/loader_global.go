package registry

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	// Ensure the wellknown types get imported and registered into the global registry
	anypb "google.golang.org/protobuf/types/known/anypb"
	apipb "google.golang.org/protobuf/types/known/apipb"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
	sourcecontextpb "google.golang.org/protobuf/types/known/sourcecontextpb"
	structpb "google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	typepb "google.golang.org/protobuf/types/known/typepb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// AddServicesFromGlobal adds the 'well known' types to the registry. This is typically implicitly called.
func AddServicesFromGlobal(registry LoaderTarget) error {
	for _, fd := range []protoreflect.FileDescriptor{
		descriptorpb.File_google_protobuf_descriptor_proto,
		anypb.File_google_protobuf_any_proto,
		apipb.File_google_protobuf_api_proto,
		durationpb.File_google_protobuf_duration_proto,
		emptypb.File_google_protobuf_empty_proto,
		fieldmaskpb.File_google_protobuf_field_mask_proto,
		sourcecontextpb.File_google_protobuf_source_context_proto,
		structpb.File_google_protobuf_struct_proto,
		timestamppb.File_google_protobuf_timestamp_proto,
		typepb.File_google_protobuf_type_proto,
		wrapperspb.File_google_protobuf_wrappers_proto,
	} {
		if err := registry.RegisterFile(fd); err != nil {
			return err
		}
	}
	return nil
}
