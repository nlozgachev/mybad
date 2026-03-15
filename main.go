// Railway-oriented error handling for Go.
// Instead of checking errors after every step, you compose a pipeline of typed transforms.
// The first error short-circuits everything that follows; `Match` forces you to handle both outcomes before you can get a value out.
package mybad
