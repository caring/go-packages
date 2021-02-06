## Pagination


This package contains the [protocol buffer][protobuf] types and associated Go  
library which are made available for use as dependencies within the Caring  
ecosystem to support Pagination within our gRPC services.

## Using these protos

In order to depend on these protos, use proto import statements that
reference the base of this repository, for example:

```protobuf
syntax = "proto3";

import "github.com/caring/go-packages/pkg/pagination/pb/pagination.proto";


// A message representing listing identities
message ListIdentityRequest {
  pagination.PaginationRequest paging = 1;
  Params params = 2;

  message Params {
    string identity_type_id = 1;
  }
}
```

If you are using `protoc` (or other similar tooling) to compile these protos yourself, 
you will require a local copy. Clone this repository to your go source (typically 
`$GOPATH/src`) and use `--proto_path` to specify this path to the compiler. You can run 
your `protoc`  command from the root directory of your project.

```bash

      PBDIR="api/pb/"
      protoc \
        --proto_path="$GOPATH/src/" \ {you may need to adjust this to your go source path}
        --proto_path=$PBDIR \
        --plugin=grpc \
        --go_out=$PBDIR --go_opt=paths=source_relative \
        --go-grpc_out=$PBDIR --go-grpc_opt=paths=source_relative \
        $PBDIR*.proto
```

## Using the go package

In order to use the package, you will need to import it into any go package
you want to use it in. Then refer to the pagination methods as needed within
your internal methods.

```go
import (
  pagination "github.com/caring/go-packages/pkg/pagination/pb"
)

type listFeatureCategoryMethods interface {
  List(context.Context, *pagination.Pager, *db.ListFeatureCategoryParams) (db.FeatureCategorySlice, *pagination.PageInfo, error)
}

// ListFeatureCategory accepts a gRPC request to list categories, executes it and returns a gRPC response
func ListFeatureCategory(ctx context.Context, req *pb.ListFeatureCategoryRequest, store listFeatureCategoryMethods) (*pb.ListFeatureCategoryResponse, error) {
  pager, err := pagination.NewPager(req.Paging)
  if err != nil {
    return nil, errors.WithGrpcStatus(err, codes.InvalidArgument)
  }

  fcs, pi, err := store.List(ctx, pager, &db.ListFeatureCategoryParams{})
  if err != nil {
    return nil, errors.WithGrpcStatus(err, codes.Internal)
  }

  return fcs.ToGraphProto(pi), nil
}
```
