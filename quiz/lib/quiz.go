package lib

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"time"
)

// QuizCmdLineArgs configures the quiz input, time limit, and question order.
type QuizCmdLineArgs struct {
	TestInputFile          string
	QuizTimeout            time.Duration
	ShouldShuffleQuestions bool
}

// QuizResults summarizes a completed or timed-out quiz.
type QuizResults struct {
	TotalNumberOfQuestions int
	NumCorrect             int
	NumIncorrect           int
}

// QuizQnA represents a question and its expected answer.
type QuizQnA struct {
	Question string
	Answer   string
}

// Run loads the configured quiz, optionally shuffles its questions, and runs it
// using in for answers and out for prompts.
func Run(args QuizCmdLineArgs, in io.Reader, out io.Writer) (QuizResults, error) {
	questionsWithAnswers, err := parseCsvFile(args.TestInputFile)
	if err != nil {
		return QuizResults{}, fmt.Errorf("error parsing %s: %w", args.TestInputFile, err)
	}
	if args.ShouldShuffleQuestions {
		shuffleQuestions(questionsWithAnswers, rand.Shuffle)
	}
	return runInteractiveQuiz(questionsWithAnswers, args.QuizTimeout, in, out)
}

func shuffleQuestions(questions []QuizQnA, shuffle func(int, func(int, int))) {
	shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})
}

func parseCsvFile(filePath string) ([]QuizQnA, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	reader := csv.NewReader(f)
	rows, readErr := reader.ReadAll()
	closeErr := f.Close()
	if readErr != nil {
		return nil, fmt.Errorf("error parsing csv: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("error closing file: %w", closeErr)
	}

	var questionsWithAnswers []QuizQnA
	for _, row := range rows {
		if len(row) != 2 {
			return nil, fmt.Errorf("invalid format for csv, each row should be two entries")
		}
		questionsWithAnswers = append(questionsWithAnswers, QuizQnA{Question: row[0], Answer: row[1]})
	}

	return questionsWithAnswers, nil
}

func runInteractiveQuiz(
	questionsWithAnswers []QuizQnA,
	timeout time.Duration,
	in io.Reader,
	out io.Writer,
) (QuizResults, error) {
	scanner := bufio.NewScanner(in)
	answerResults := make(chan bool)
	done := make(chan struct{})

	totalNumberOfQuestions := len(questionsWithAnswers)
	var numCorrect int

	if totalNumberOfQuestions == 0 {
		return QuizResults{}, nil
	}

	fmt.Fprint(out, "Press Enter to start the quiz...")
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return QuizResults{}, fmt.Errorf("error reading input: %w", err)
		}
		return QuizResults{
			TotalNumberOfQuestions: totalNumberOfQuestions,
			NumIncorrect:           totalNumberOfQuestions,
		}, nil
	}

	go func() {
		defer close(answerResults)
		for i, question := range questionsWithAnswers {
			_, _ = fmt.Fprintf(out, "Question %d: %s=", i+1, question.Question)
			if !scanner.Scan() {
				return
			}
			answer := strings.TrimSpace(scanner.Text())
			isCorrect := answer == question.Answer

			select {
			case answerResults <- isCorrect:
				// Result delivered successfully.
			case <-done:
				return
			}
		}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	defer close(done)

	for {
		select {
		case <-timer.C:
			fmt.Fprintln(out)
			return QuizResults{
				TotalNumberOfQuestions: totalNumberOfQuestions,
				NumCorrect:             numCorrect,
				NumIncorrect:           totalNumberOfQuestions - numCorrect,
			}, nil

		case isCorrect, ok := <-answerResults:
			if !ok {
				if err := scanner.Err(); err != nil {
					return QuizResults{},
						fmt.Errorf("error reading input: %w", err)
				}

				return QuizResults{
					TotalNumberOfQuestions: totalNumberOfQuestions,
					NumCorrect:             numCorrect,
					NumIncorrect:           totalNumberOfQuestions - numCorrect,
				}, nil
			}

			if isCorrect {
				numCorrect++
			}
		}
	}
}
