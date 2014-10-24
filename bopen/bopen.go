package bopen

type Opener interface {
	Open(url string)
}

type Bopen struct {
	Topic string
}

func NewOpener(topic string) Opener {
	return &Bopen{topic}
}
