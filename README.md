YEMA - the interchangeable schema language
------------------------------------------

yema is extremly simple and obvious.
its only purpose is to generate type bindings for other languages,
it does intentionally NOT support constraints like jsonschema.

yema can be defined in yaml or json or whatever else.
here's a weird yaml to showcase all of it:

```yaml
firstName:        string
middleName?:      string
lastName:         string
age:              int32
favoriteNums?:    [int]
phones:           [string]
addresses: [{
  street:     string,
  city:       string,
  zipCode:    string
}]
favoriteObjects:
  - color: string
    shape?: int
settings: 
  notifications:  bool
  theme:          string
  limits?: 
    max:          int64
```


you can use it as cli to generate types:

    go install github.com/aep/yema/cmd/yema@latest

    yema -o cue < yema.yaml
