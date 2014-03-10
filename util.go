package main

func urisFromOpts(opts GlobalOptions) map[string]string {

	uris := make(map[string]string)
	uris["primary"] = opts.Primary
	uris["introducer"] = opts.Introducer

	return uris
}
