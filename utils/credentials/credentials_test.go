package credentials

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func TestTwiceAttachToContext(t *testing.T) {
	authorizedUser = userCredentials{
		username: "foo",
		password: "bar",
	}
	ctx := AttachToContext(context.Background())
	ctx = AttachToContext(ctx)
	got, _ := metadata.FromOutgoingContext(ctx)
	want := metadata.MD{
		"username": []string{"foo"},
		"password": []string{"bar"},
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("(-got, +want):\n%s", diff)
	}
}
