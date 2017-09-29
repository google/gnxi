package xpath

import (
	"reflect"
	"testing"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

func TestToGNMIPath(t *testing.T) {
	tests := []struct {
		desc     string
		path     string
		expectOK bool
		want     *pb.Path
	}{{
		desc:     "empty path",
		path:     "",
		expectOK: true,
		want:     &pb.Path{},
	}, {
		desc:     "root path",
		path:     "/",
		expectOK: true,
		want:     &pb.Path{},
	}, {
		desc:     "test path with root omitted",
		path:     "a",
		expectOK: true,
		want: &pb.Path{
			Elem: []*pb.PathElem{
				&pb.PathElem{Name: "a"},
			},
		},
	}, {
		desc:     "test path with trailing / separator",
		path:     "/a/",
		expectOK: true,
		want: &pb.Path{
			Elem: []*pb.PathElem{
				&pb.PathElem{Name: "a"},
			},
		},
	}, {
		desc:     "test path with mutiple / separators",
		path:     "/a//b",
		expectOK: true,
		want: &pb.Path{
			Elem: []*pb.PathElem{
				&pb.PathElem{Name: "a"},
				&pb.PathElem{Name: "b"},
			},
		},
	}, {
		desc:     "test path without attribute",
		path:     "/a/b/c",
		expectOK: true,
		want: &pb.Path{
			Elem: []*pb.PathElem{
				&pb.PathElem{Name: "a"},
				&pb.PathElem{Name: "b"},
				&pb.PathElem{Name: "c"},
			},
		},
	}, {
		desc:     "test path with one attribute",
		path:     "/a/b[k=10]/c",
		expectOK: true,
		want: &pb.Path{
			Elem: []*pb.PathElem{
				&pb.PathElem{Name: "a"},
				&pb.PathElem{Name: "b", Key: map[string]string{"k": "10"}},
				&pb.PathElem{Name: "c"},
			},
		},
	}, {
		desc:     "test path with multiple attributes",
		path:     "/a/b[k1=10][k2=10.10.10.10/24]/c",
		expectOK: true,
		want: &pb.Path{
			Elem: []*pb.PathElem{
				&pb.PathElem{Name: "a"},
				&pb.PathElem{Name: "b", Key: map[string]string{"k1": "10", "k2": "10.10.10.10/24"}},
				&pb.PathElem{Name: "c"},
			},
		},
	}, {
		desc:     "test path without name",
		path:     "/[k1=10][k2=20]/c",
		expectOK: false,
		want:     nil,
	}, {
		desc:     "subsequent test paths without name",
		path:     "/[k1=10]/[k2=20]/c",
		expectOK: false,
		want:     nil,
	}}

	for _, test := range tests {
		got, err := ToGNMIPath(test.path)
		if test.expectOK {
			if err != nil {
				t.Errorf("%s: ToGNMIPath(%q) got error: %v, wanted error: %v",
					test.desc, test.path, err, !test.expectOK)
				continue
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("%s: ToGNMIPath(%q) got: %v, wanted: %v",
					test.desc, test.path, got, test.want)
			}
		} else if err == nil {
			t.Errorf("%s: ToGNMIPath(%q) got error: %v, wanted error: %v",
				test.desc, test.path, err, !test.expectOK)
		}
	}
}
