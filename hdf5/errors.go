package hdf5

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

var (
	ErrClosedGroup      = errors.New("hdf5: group is closed")
	ErrClosedFile       = errors.New("hdf5: file is closed")
	ErrNotFound         = errors.New("hdf5: object not found")
	ErrInvalidDatatype  = errors.New("hdf5: invalid datatype")
	ErrInvalidDataspace = errors.New("hdf5: invalid dataspace")
	ErrInvalidLayout    = errors.New("hdf5: invalid layout")
	ErrReadOnly         = errors.New("hdf5: file is read-only")
	ErrWriteOnly        = errors.New("hdf5: write-only")
	ErrUnsupported      = errors.New("hdf5: operation not supported")
	ErrInternal         = errors.New("hdf5: internal error")
	ErrExternalLink     = errors.New("hdf5: external links are not supported")
	ErrInvalidSelection = errors.New("hdf5: invalid selection")
	ErrInvalidArgument  = errors.New("hdf5: invalid argument")
	ErrIO               = errors.New("hdf5: I/O error")
	ErrMemory           = errors.New("hdf5: memory allocation error")
	ErrTimeout          = errors.New("hdf5: timeout")
	ErrDeadlock         = errors.New("hdf5: deadlock detected")
	ErrBusy             = errors.New("hdf5: resource is busy")
	ErrCorrupt          = errors.New("hdf5: file is corrupt")
	ErrIncompatible     = errors.New("hdf5: incompatible format")
	ErrOverflow         = errors.New("hdf5: overflow")
	ErrUnderflow        = errors.New("hdf5: underflow")
)

type ErrorCode int

const (
	ErrorCodeOK             ErrorCode = 0
	ErrorCodeError          ErrorCode = -1
	ErrorCodeBadID          ErrorCode = -2
	ErrorCodeBadParam       ErrorCode = -3
	ErrorCodeBadType        ErrorCode = -4
	ErrorCodeBadAlloc       ErrorCode = -5
	ErrorCodeBusy           ErrorCode = -6
	ErrorCodeCantOpenFile   ErrorCode = -7
	ErrorCodeCantCreateFile ErrorCode = -8
	ErrorCodeCantRead       ErrorCode = -9
	ErrorCodeCantWrite      ErrorCode = -10
	ErrorCodeCantSeek       ErrorCode = -11
	ErrorCodeEndOfFile      ErrorCode = -12
	ErrorCodeCorruptFile    ErrorCode = -13
	ErrorCodeUnsupported    ErrorCode = -14
	ErrorCodeOverflow       ErrorCode = -15
	ErrorCodeUnderflow      ErrorCode = -16
)

type HDF5Error struct {
	Code     ErrorCode
	Message  string
	Location string
	Stack    []string
	Cause    error
}

func NewError(code ErrorCode, message string) *HDF5Error {
	return &HDF5Error{
		Code:     code,
		Message:  message,
		Location: getCallerLocation(),
		Stack:    getStackTrace(),
	}
}

func NewErrorWithCause(code ErrorCode, message string, cause error) *HDF5Error {
	return &HDF5Error{
		Code:     code,
		Message:  message,
		Location: getCallerLocation(),
		Stack:    getStackTrace(),
		Cause:    cause,
	}
}

func (e *HDF5Error) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[HDF5 Error %d] %s", e.Code, e.Message))
	if e.Location != "" {
		sb.WriteString(fmt.Sprintf(" at %s", e.Location))
	}
	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf(" (caused by: %v)", e.Cause))
	}
	return sb.String()
}

func (e *HDF5Error) Unwrap() error {
	return e.Cause
}

func (e *HDF5Error) Format(f fmt.State, c rune) {
	switch c {
	case 'v':
		if f.Flag('+') {
			fmt.Fprintf(f, "[HDF5 Error %d] %s\n", e.Code, e.Message)
			if e.Location != "" {
				fmt.Fprintf(f, "Location: %s\n", e.Location)
			}
			if len(e.Stack) > 0 {
				fmt.Fprintf(f, "Stack trace:\n")
				for _, frame := range e.Stack {
					fmt.Fprintf(f, "  %s\n", frame)
				}
			}
			if e.Cause != nil {
				fmt.Fprintf(f, "Cause: %v\n", e.Cause)
			}
			return
		}
		fallthrough
	default:
		fmt.Fprint(f, e.Error())
	}
}

func getCallerLocation() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func getStackTrace() []string {
	var stack []string
	for i := 4; i < 10; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stack = append(stack, fmt.Sprintf("%s:%d", file, line))
	}
	return stack
}

func IsHDF5Error(err error) bool {
	var hdf5Err *HDF5Error
	return errors.As(err, &hdf5Err)
}

func GetErrorCode(err error) ErrorCode {
	var hdf5Err *HDF5Error
	if errors.As(err, &hdf5Err) {
		return hdf5Err.Code
	}
	return ErrorCodeError
}

func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	var hdf5Err *HDF5Error
	if errors.As(err, &hdf5Err) {
		return NewErrorWithCause(hdf5Err.Code, message+": "+hdf5Err.Message, err)
	}
	return NewErrorWithCause(ErrorCodeError, message, err)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}


