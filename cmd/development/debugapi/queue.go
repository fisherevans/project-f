package debugapi

import (
    "context"
    "net/http"
)

type commandRequest struct {
    fn     func() (any, error)
    result chan commandResult
}

type commandResult struct {
    value any
    err   error
}

type CommandQueue struct {
    ch chan commandRequest
}

func NewCommandQueue(bufferSize int) *CommandQueue {
    return &CommandQueue{
        ch: make(chan commandRequest, bufferSize),
    }
}

func (q *CommandQueue) Enqueue(ctx context.Context, fn func() (any, error)) (any, error) {
    req := commandRequest{
        fn:     fn,
        result: make(chan commandResult, 1),
    }
    select {
    case q.ch <- req:
    case <-ctx.Done():
        return nil, ctx.Err()
    }
    select {
    case res := <-req.result:
        return res.value, res.err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (q *CommandQueue) DrainOnGameThread() {
    for {
        select {
        case req := <-q.ch:
            v, err := req.fn()
            req.result <- commandResult{value: v, err: err}
        default:
            return
        }
    }
}

func (q *CommandQueue) handleOnGameThread(w http.ResponseWriter, r *http.Request, fn func() (any, error)) {
    result, err := q.Enqueue(r.Context(), fn)
    if err == context.DeadlineExceeded || err == context.Canceled {
        http.Error(w, "game loop timeout", http.StatusServiceUnavailable)
        return
    }
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    if result == nil {
        w.WriteHeader(http.StatusNoContent)
        return
    }
    writeJSON(w, http.StatusOK, result)
}
