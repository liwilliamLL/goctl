package generator

import (
	"fmt"
	proto2 "github.com/emicklei/proto"
	"log"
	"path/filepath"

	conf "william/base/go-zero/tools/goctl/config"
	"william/base/go-zero/tools/goctl/rpc/parser"
	"william/base/go-zero/tools/goctl/util"
	"william/base/go-zero/tools/goctl/util/console"
	"william/base/go-zero/tools/goctl/util/ctx"
)

// RPCGenerator defines a generator and configure
type RPCGenerator struct {
	g   Generator
	cfg *conf.Config
}

// NewDefaultRPCGenerator wraps Generator with configure
func NewDefaultRPCGenerator(style string, experimental_allow_proto3_optional bool) (*RPCGenerator, error) {
	cfg, err := conf.NewConfig(style, experimental_allow_proto3_optional)
	if err != nil {
		return nil, err
	}
	return NewRPCGenerator(NewDefaultGenerator(), cfg), nil
}

// NewRPCGenerator creates an instance for RPCGenerator
func NewRPCGenerator(g Generator, cfg *conf.Config) *RPCGenerator {
	return &RPCGenerator{
		g:   g,
		cfg: cfg,
	}
}

// Generate generates an rpc service, through the proto file,
// code storage directory, and proto import parameters to control
// the source file and target location of the rpc service that needs to be generated
func (g *RPCGenerator) Generate(src, target string, protoImportPath []string, output, callo string, only_client bool) error {

	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	err = util.MkdirIfNotExist(abs)
	if err != nil {
		return err
	}

	err = g.g.Prepare()
	if err != nil {
		return err
	}

	projectCtx, err := ctx.Prepare(abs)
	if err != nil {
		return err
	}

	p := parser.NewDefaultProtoParser()
	proto, err := p.Parse(src)
	if err != nil {
		return err
	}

	dirCtx, err := mkdir(projectCtx, !only_client, proto, output, callo)
	if err != nil {
		return err
	}

	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	srcDir := filepath.Dir(absSrc)
	for _, path := range protoImportPath {
		for _, imp := range proto.Import {
			for _, src := range []string{
				fmt.Sprintf("%s/%s", path, imp.Import.Filename),
				fmt.Sprintf("%s/%s", srcDir, imp.Import.Filename),
				imp.Import.Filename} {
				importProto, _ := p.Parse(src)
				if &importProto != nil && importProto.Message != nil {

					importProto.Name = imp.Import.Filename
					if path != srcDir {
						importProto.Src = absSrc
					}

					err = g.g.GenPb(dirCtx, protoImportPath, importProto, g.cfg)
					if err != nil {
						log.Println("err if GenPb")
						return err
					}

					for _, message := range importProto.Message {
						proto.ImportMessage = append(proto.ImportMessage, message)
					}
				}
			}
		}
	}

	for _, element := range proto.Service.Service.Elements {
		rpc, ok := element.(*proto2.RPC)
		if !ok {
			continue
		}

		lt := searchMessage(rpc.RequestType, proto.Message)
		if lt == nil {
			// 在当前proto文件里没有找到左值，去import包里找
			t := searchMessage(rpc.RequestType, proto.ImportMessage)
			if t != nil {
				proto.Message = append(proto.Message, *t)
			}
		}

		rt := searchMessage(rpc.ReturnsType, proto.Message)
		if rt == nil {
			// 在当前proto文件里没有找到右值，去import包里找
			t := searchMessage(rpc.ReturnsType, proto.ImportMessage)
			if t != nil {
				proto.Message = append(proto.Message, *t)
			}
		}
	}

	if !only_client {

		err = g.g.GenEtc(dirCtx, proto, g.cfg)
		if err != nil {
			log.Println("err if GenEtc")
			return err
		}

		err = g.g.GenConfig(dirCtx, proto, g.cfg)
		if err != nil {
			log.Println("err if GenConfig")
			return err
		}

		err = g.g.GenSvc(dirCtx, proto, g.cfg)
		if err != nil {
			log.Println("err if GenSvc")

			return err
		}

		err = g.g.GenLogic(dirCtx, proto, g.cfg)
		if err != nil {
			log.Println("err if GenLogic")
			return err
		}

		err = g.g.GenServer(dirCtx, proto, g.cfg)
		if err != nil {
			log.Println("err if GenServer")
			return err
		}

		err = g.g.GenMain(dirCtx, proto, g.cfg)
		if err != nil {
			log.Println("err if GenMain")
			return err
		}

		err = g.g.GenHttp(dirCtx, proto, g.cfg)
		if err != nil {
			log.Println("err if GenHttp")
			return err
		}
	}

	err = g.g.GenPb(dirCtx, protoImportPath, proto, g.cfg)
	if err != nil {
		log.Println("err if GenPb")
		return err
	}

	err = g.g.GenCall(dirCtx, proto, g.cfg)
	if err != nil {
		log.Println("err if GenCall")
		return err
	}

	console.NewColorConsole().MarkDone()

	return err
}

func searchMessage(t string, msgs []parser.Message) (msg *parser.Message) {

	for _, m := range msgs {
		if m.Name == t {
			return &m
		}
	}
	return
}
