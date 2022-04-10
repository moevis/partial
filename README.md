# Partial - generating partial structs for golang

Define a single struct, and then split it into multiple.

## Installation

```shell
go install github.com/moevis/partial
```

## Usage

Using tag "partial" and "+/-" to include or exclude field. For example, if you have a `models.go` file like following code.

```golang
package models

//go:generate partial -type=Person $GOFILE

type Person struct {
  // define a struct naming PersonWithName with a single field - "Name"
  Name     string `partial:"PersonWithName,PersonWithNameAndAge"`
  Age      int    `partial:"PersonWithNameAndAge"`
  // define a struct without Password field
  Password string `partial:"-NoPasswordPerson"`
}
```

Run command:

```shell
go generate
```

And then you will get three files in the same folder of models.go: 
- personwithname.go
- personwithnameandage.go
- nopasswordperson.go



Content of personwithname.go

```golang
package models

type PersonWithName struct {
  Name string
}

```

Content of personwithnameandage.go

```golang
package models

//go:generate partial -type=Person $GOFILE

type Person struct {
  // define a struct naming PersonWithName with a single field - "Name"
  Name     string `partial:"PersonWithName,PersonWithNameAndAge"`
  Age      int    `partial:"PersonWithNameAndAge"`
}
```
