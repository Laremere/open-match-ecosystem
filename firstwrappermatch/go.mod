module open-match.dev/open-match-ecosystem/firstwrappermatch

go 1.13

require (
	github.com/sirupsen/logrus v1.4.2
	google.golang.org/grpc v1.25.1
	open-match.dev/open-match v0.8.0
	open-match.dev/open-match-ecosystem/demoui v0.0.0-local
	open-match.dev/open-match-ecosystem/wrapper v0.0.0-local
)

replace open-match.dev/open-match-ecosystem/demoui v0.0.0-local => ../demoui

replace open-match.dev/open-match-ecosystem/wrapper v0.0.0-local => ../wrapper
