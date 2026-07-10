package lib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const ownerReadWrite os.FileMode = 0o600

func TestRun(t *testing.T) {
	tests := []struct {
		name          string
		csv           string
		input         string
		timeout       time.Duration
		keepInputOpen bool
		skipCSVWrite  bool
		want          QuizResults
		wantErr       string
	}{
		{
			name:    "all answers correct",
			csv:     "2+2,4\n3+5,8\n",
			input:   "\n4\n8\n",
			timeout: time.Second,
			want:    QuizResults{TotalNumberOfQuestions: 2, NumCorrect: 2},
		},
		{
			name:    "mixture of correct and incorrect answers",
			csv:     "2+2,4\n3+5,8\n",
			input:   "\n4\n7\n",
			timeout: time.Second,
			want:    QuizResults{TotalNumberOfQuestions: 2, NumCorrect: 1, NumIncorrect: 1},
		},
		{
			name:    "answers are trimmed",
			csv:     "2+2,4\n",
			input:   "\n  4  \n",
			timeout: time.Second,
			want:    QuizResults{TotalNumberOfQuestions: 1, NumCorrect: 1},
		},
		{
			name:    "quoted question containing a comma",
			csv:     "\"what 2+2, sir?\",4\n",
			input:   "\n4\n",
			timeout: time.Second,
			want:    QuizResults{TotalNumberOfQuestions: 1, NumCorrect: 1},
		},
		{
			name:    "empty lib",
			timeout: time.Second,
			want:    QuizResults{},
		},
		{
			name:    "EOF counts unanswered questions as incorrect",
			csv:     "2+2,4\n3+5,8\n",
			input:   "\n4\n",
			timeout: time.Second,
			want:    QuizResults{TotalNumberOfQuestions: 2, NumCorrect: 1, NumIncorrect: 1},
		},
		{
			name:          "timeout counts unanswered questions as incorrect",
			csv:           "2+2,4\n3+5,8\n",
			input:         "\n",
			timeout:       100 * time.Millisecond,
			keepInputOpen: true,
			want:          QuizResults{TotalNumberOfQuestions: 2, NumIncorrect: 2},
		},
		{
			name:         "missing CSV file",
			timeout:      time.Second,
			skipCSVWrite: true,
			wantErr:      "error opening file",
		},
		{
			name:    "malformed CSV",
			csv:     "\"unterminated",
			timeout: time.Second,
			wantErr: "error parsing csv",
		},
		{
			name:    "row has too few fields",
			csv:     "2+2\n",
			timeout: time.Second,
			wantErr: "each row should be two entries",
		},
		{
			name:    "row has too many fields",
			csv:     "2+2,4,extra\n",
			timeout: time.Second,
			wantErr: "each row should be two entries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvPath := filepath.Join(t.TempDir(), "problems.csv")
			if !tt.skipCSVWrite {
				if err := os.WriteFile(csvPath, []byte(tt.csv), ownerReadWrite); err != nil {
					t.Fatalf("write CSV: %v", err)
				}
			}

			var input io.Reader = strings.NewReader(tt.input)
			if tt.keepInputOpen {
				reader, writer := io.Pipe()
				input = reader
				t.Cleanup(func() {
					_ = reader.Close()
					_ = writer.Close()
				})

				go func() {
					_, _ = fmt.Fprint(writer, tt.input)
				}()
			}

			got, err := Run(QuizCmdLineArgs{
				TestInputFile: csvPath,
				QuizTimeout:   tt.timeout,
			}, input, io.Discard)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Run() error = nil; want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Run() error = %q; want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Run() returned unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Run() = %+v; want %+v", got, tt.want)
			}
		})
	}
}

func TestShuffleQuestions(t *testing.T) {
	questions := []QuizQnA{
		{Question: "first", Answer: "1"},
		{Question: "second", Answer: "2"},
		{Question: "third", Answer: "3"},
	}

	reverse := func(n int, swap func(int, int)) {
		for i := 0; i < n/2; i++ {
			swap(i, n-1-i)
		}
	}

	shuffleQuestions(questions, reverse)

	want := []QuizQnA{
		{Question: "third", Answer: "3"},
		{Question: "second", Answer: "2"},
		{Question: "first", Answer: "1"},
	}
	for i := range want {
		if questions[i] != want[i] {
			t.Fatalf("shuffleQuestions() = %+v; want %+v", questions, want)
		}
	}
}
