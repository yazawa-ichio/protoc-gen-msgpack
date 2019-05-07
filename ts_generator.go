package main

import (
	"strings"

	"github.com/golang/protobuf/proto"
	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pseudomuto/protokit"
	ce "github.com/yazawa-ichio/protoc-gen-msgpack/code_emitter"
)

type tsGenerator struct {
	e    *ce.CodeEmitter
	conv *TSConverter
}

func newTSGenerator() *tsGenerator {
	return &tsGenerator{e: new(ce.CodeEmitter), conv: &TSConverter{}}
}

func (g *tsGenerator) genResponseFile(data *protoData) []*plugin_go.CodeGeneratorResponse_File {
	files := []*plugin_go.CodeGeneratorResponse_File{}
	for _, msg := range data.messages {
		files = append(files, g.genClass(msg))
	}
	for _, enum := range data.enums {
		files = append(files, g.genEnum(enum))
	}
	return files
}

func (g *tsGenerator) genClass(message *messageData) *plugin_go.CodeGeneratorResponse_File {
	g.e.Reset()
	emitFileInfo(g.e, message.file)
	g.emitDeps(message.data)
	g.emitClass(message)
	fileName := g.conv.GetFileName(message.data.GetFullName())
	content := g.e.String()
	return &plugin_go.CodeGeneratorResponse_File{
		Name:    proto.String(fileName),
		Content: proto.String(content),
	}
}

func (g *tsGenerator) getRequireName(typeName string) string {
	name := g.conv.GetFileName(typeName)
	return strings.Replace(strings.Replace(name, ".", "_", -1), "-", "$", -1)
}

func (g *tsGenerator) emitDeps(m *protokit.Descriptor) {
	g.e.EmitLine("/// <reference types=\"node\" />")
	g.e.EmitLine("import * as packer from 'proto-msgpack'")
	hits := make(map[string]string)
	for _, f := range m.GetMessageFields() {
		if !isUserDefine(f) {
			continue
		}
		name := f.GetTypeName()
		if _, hit := hits[f.GetTypeName()]; hit {
			continue
		}
		hits[name] = name
		g.e.EmitLine("import %s = require('./%s');", g.importName(name), g.formFileName(name))
	}
}

func (g *tsGenerator) formFileName(name string) string {
	name = g.conv.GetFileName(name)
	name = name[:len(name)-5]
	return name
}

func (g *tsGenerator) importName(name string) string {
	return strings.Replace(g.formFileName(name), ".", "_", -1)
}

func (g *tsGenerator) emitClass(message *messageData) {
	g.emitComment(message.data.GetComments().GetLeading())
	g.e.StartBracket("declare class %s", message.data.GetName())
	for _, f := range message.data.GetMessageFields() {
		typeName := g.conv.GetTypeImpl(f)
		if isUserDefine(f) {
			typeName = g.importName(typeName)
		}
		if f.GetType().String() == "TYPE_MESSAGE" || f.GetType().String() == "TYPE_BYTES" {
			typeName = typeName + " | null"
		}
		if f.GetLabel().String() == "LABEL_REPEATED" {
			typeName = "Array<" + typeName + "> | null"
		}
		g.e.EmitLine("%s: %s;", f.GetName(), typeName)
	}
	g.e.EmitLine("constructor(init?: boolean | Buffer, pos?: number) ")
	g.e.EmitLine("pack(): Buffer;")
	g.e.EmitLine("unpack(buf: Buffer, pos?: number): void;")
	g.e.EmitLine("write(w: packer.ProtoWriter): void;")
	g.e.EmitLine("read(r: packer.ProtoReader): void;")
	g.e.EndBracket("")
	g.e.EmitLine("export = " + message.data.GetName() + ";")
}

func (g *tsGenerator) genEnum(enum *enumData) *plugin_go.CodeGeneratorResponse_File {
	g.e.Reset()
	emitFileInfo(g.e, enum.file)
	g.emitEnum(enum)
	fileName := g.conv.GetFileName(enum.data.GetFullName())
	content := g.e.String()
	return &plugin_go.CodeGeneratorResponse_File{
		Name:    proto.String(fileName),
		Content: proto.String(content),
	}
}

func (g *tsGenerator) emitEnum(enum *enumData) {
	g.emitComment(enum.data.GetComments().GetLeading())
	g.e.Bracket("declare enum %s", func() {
		vals := enum.data.GetValues()
		for i, v := range vals {
			g.emitComment(v.GetComments().GetLeading())
			if i < len(vals)-1 {
				g.e.EmitLine("%s = %d,", g.conv.GetEnumName(v.GetName()), v.GetNumber())
			} else {
				g.e.EmitLine("%s = %d", g.conv.GetEnumName(v.GetName()), v.GetNumber())
			}
		}
	}, enum.data.GetName())
	g.emitComment(enum.data.GetComments().GetLeading())
	g.e.EmitLine("export = %s;", enum.data.GetName())
}

func (g *tsGenerator) emitComment(comment string) bool {
	if comment == "" {
		return false
	}
	g.e.EmitLine("/*")
	for _, c := range strings.Split(comment, "\n") {
		g.e.EmitLine(c)
	}
	g.e.EmitLine("*/")
	return true
}