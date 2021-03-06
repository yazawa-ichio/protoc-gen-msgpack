// tests/proto/import.proto
"use strict";
const _packer = require('proto-msgpack');
const _proto = require('./index.js');
class DependMessage {
	constructor(init, pos) {
		this.text = null;
		if(Buffer.isBuffer(init)){
			this.read(new _packer.ProtoReader(init, pos));
		}
	}
	pack() {
		const w = _packer.defaultWriter;
		w.clear();
		this.write(w);
		return w.toBuffer();
	}
	unpack(buf, pos) {
		if(!Buffer.isBuffer(buf)){
			this.read(buf);
		} else {
			this.read(new _packer.ProtoReader(buf, pos));
		}
	}
	write(w) {
		// Write Map Length
		w.writeMapHeader(1);
		
		// Write text
		w.writeTag(500);
		w.writeString(this.text);
	}
	read(r) {
		// Read Map Length
		const mapLen = r.readMapHeader();

		for(let i = 0; i < mapLen; i++) {
			const tag = r.readTag();
			switch(tag) {
			case 500:
				this.text = r.readString();
				break;
			default:
				r.skip();
				break;
			}
		}
	}
}
module.exports = DependMessage
