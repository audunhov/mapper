# mapper

mapper is a Go library for importing CSV and XLSX files and mapping the data to Go structs.

## Installation

```bash
go get github.com/audunhov/mapper
```

## Example

```go
package main

import (
	"fmt"
	"strings"

	"github.com/audunhov/mapper"
)

type User struct {
	FirstName string `map:"First Name"`
	LastName  string
	Age       int
	IsActive  bool   `map:"active"`
}

func main() {
	csvData := `First Name,last_name,Age,active
Alice,Smith,30,true
Bob,Jones,25,false`

	importer := mapper.NewCSVImporter(strings.NewReader(csvData))
	imported, err := importer.Import()
	if err != nil {
		panic(err)
	}

	users, err := mapper.Convert[User](imported, nil)
	if err != nil {
		panic(err)
	}

	for _, u := range users {
		fmt.Printf("%s %s (Age: %d, Active: %t)\n", u.FirstName, u.LastName, u.Age, u.IsActive)
	}
}
```

## Documentation

See [examples.md](file:///home/audun/code/mapper/examples.md) and [web_example.md](file:///home/audun/code/mapper/web_example.md) for more details.

## License

MIT
