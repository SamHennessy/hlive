package hlive

import "testing"

func Test_pathGreater(t *testing.T) {
	type args struct {
		pathA string
		pathB string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"both empty",
			args{
				pathA: "",
				pathB: "",
			},
			false,
		},
		{
			"one level equal",
			args{
				pathA: "0",
				pathB: "0",
			},
			false,
		},
		{
			"one level greater",
			args{
				pathA: "1",
				pathB: "0",
			},
			true,
		},
		{
			"one level less",
			args{
				pathA: "0",
				pathB: "1",
			},
			false,
		},
		{
			"two levels a equal",
			args{
				pathA: "0>1",
				pathB: "0>1",
			},
			false,
		},
		{
			"two levels a greater",
			args{
				pathA: "0>2",
				pathB: "0>1",
			},
			true,
		},
		{
			"two levels a less",
			args{
				pathA: "0>1",
				pathB: "0>2",
			},
			false,
		},
		{
			"three levels a equal",
			args{
				pathA: "0>1>2",
				pathB: "0>1>2",
			},
			false,
		},
		{
			"three levels a greater",
			args{
				pathA: "0>1>2",
				pathB: "0>1>1",
			},
			true,
		},
		{
			"three levels a less",
			args{
				pathA: "0>1>1",
				pathB: "0>1>2",
			},
			false,
		},
		{
			"mid path diff greater",
			args{
				pathA: "0>2>1",
				pathB: "0>1>1",
			},
			true,
		},
		{
			"mid path diff less",
			args{
				pathA: "0>1>1",
				pathB: "0>2>1",
			},
			false,
		},
		{
			"mis match length less",
			args{
				pathA: "0>1>2>3",
				pathB: "0>1>2",
			},
			false,
		},
		{
			"mis match length greater",
			args{
				pathA: "0>1>2",
				pathB: "0>1>2>3",
			},
			false,
		},
		{
			"real example",
			args{
				pathA: "1>1>2>1",
				pathB: "1>1>2>2",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathGreater(tt.args.pathA, tt.args.pathB); got != tt.want {
				t.Errorf("pathGreater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pathLesser(t *testing.T) {
	type args struct {
		pathA string
		pathB string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"both empty",
			args{
				pathA: "",
				pathB: "",
			},
			false,
		},
		{
			"one level equal",
			args{
				pathA: "0",
				pathB: "0",
			},
			false,
		},
		{
			"one level greater",
			args{
				pathA: "1",
				pathB: "0",
			},
			false,
		},
		{
			"one level less",
			args{
				pathA: "0",
				pathB: "1",
			},
			true,
		},
		{
			"two levels a equal",
			args{
				pathA: "0>1",
				pathB: "0>1",
			},
			false,
		},
		{
			"two levels a greater",
			args{
				pathA: "0>2",
				pathB: "0>1",
			},
			false,
		},
		{
			"two levels a less",
			args{
				pathA: "0>1",
				pathB: "0>2",
			},
			true,
		},
		{
			"three levels a equal",
			args{
				pathA: "0>1>2",
				pathB: "0>1>2",
			},
			false,
		},
		{
			"three levels a greater",
			args{
				pathA: "0>1>2",
				pathB: "0>1>1",
			},
			false,
		},
		{
			"three levels a less",
			args{
				pathA: "0>1>1",
				pathB: "0>1>2",
			},
			true,
		},
		{
			"mid path diff greater",
			args{
				pathA: "0>2>1",
				pathB: "0>1>1",
			},
			false,
		},
		{
			"mid path diff less",
			args{
				pathA: "0>1>1",
				pathB: "0>2>1",
			},
			true,
		},
		{
			"mis match length less",
			args{
				pathA: "0>1>2>3",
				pathB: "0>1>2",
			},
			true,
		},
		{
			"mis match length greater",
			args{
				pathA: "0>1>2",
				pathB: "0>1>2>3",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathLesser(tt.args.pathA, tt.args.pathB); got != tt.want {
				t.Errorf("pathLesser() = %v, want %v", got, tt.want)
			}
		})
	}
}
