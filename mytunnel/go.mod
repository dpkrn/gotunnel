module github.com/dpkrn/gotunnel/mytunnel

go 1.25

require github.com/dpkrn/gotunnel v0.3.6

// Parent module path must be lowercase for go get (github.com/dpkrn/...).
// Remove this replace after you tag a release whose go.mod uses github.com/dpkrn/gotunnel.
replace github.com/dpkrn/gotunnel => ../

require github.com/hashicorp/yamux v0.1.2 // indirect
