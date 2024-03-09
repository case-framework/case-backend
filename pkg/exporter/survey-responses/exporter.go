package surveyresponses

import (
	"fmt"
	"io"
)

type ResponseExporter struct {
	parser *ResponseParser
	writer *io.Writer
	format string
}

func NewResponseExporter(
	parser *ResponseParser,
	writer *io.Writer,
	format string,
) (*ResponseExporter, error) {
	re := &ResponseExporter{
		parser: parser,
		writer: writer,
		format: format,
	}

	if err := re.init(); err != nil {
		return nil, err
	}

	return re, nil
}

func (re *ResponseExporter) init() error {
	// depending on format, init writer
	// if csv write header
	// if json write [{ "responses": [              ...
	return nil
}

func (re *ResponseExporter) WriteResponse(
	parsedResponse ParsedResponse,
) error {
	switch re.format {
	case "csv":
		// write to csv
	case "json":
		// write to json
	default:
		return fmt.Errorf("unsupported format: %s", re.format)
	}
	return nil
}

func (re *ResponseExporter) Finish() error {

	// if json write ] }

	return nil
}

// TODO: method that will init file CSV or JSON -> use parser and receive writer
// TODO: method that will write to file CSV or JSON -> use parser and receive writer and new response to be processed
// TODO: method that will close file CSV call flush or JSON -> use parser and receive writer
