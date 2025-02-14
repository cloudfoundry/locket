set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "$DIR/../models"
protoc --proto_path=../../vendor:../../vendor/google.golang.org/protobuf/types/known/durationpb/:. \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    ./*.proto
popd
