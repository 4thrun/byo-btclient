# Build My Own [BitTorrent Client](https://github.com/codecrafters-io/build-your-own-x?tab=readme-ov-file#build-your-own-bittorrent-client) in Go 

## Reference

- [Building a BitTorrent client from the ground up in Go](https://blog.jse.li/posts/torrent/)
- [BitTorrentSpecification](https://wiki.theory.org/BitTorrentSpecification)
- [veggiedefender/torrent-client: Tiny BitTorrent client written in Go](https://github.com/veggiedefender/torrent-client)

## Format 

### .torrent Example 

```
d
  8:announce
    41:http://bttracker.debian.org:6969/announce
  7:comment
    35:"Debian CD from cdimage.debian.org"
  13:creation date
    i1573903810e
  4:info
    d
      6:length
        i351272960e
      4:name
        31:debian-10.2.0-amd64-netinst.iso
      12:piece length
        i262144e
      6:pieces
        26800:�����PS�^�� (binary blob of the hashes of each piece)
    e
e
```

### Tracker Response Example 

```
d
  8:interval
    i900e
  5:peers
    252:(another long binary blob)
e
```

### Peer Handshake Example 

```
\x13BitTorrent protocol\x00\x00\x00\x00\x00\x00\x00\x00\x86\xd4\xc8\x00\x24\xa4\x69\xbe\x4c\x50\xbc\x5a\x10\x2c\xf7\x17\x80\x31\x00\x74-TR2940-k8hj0wgej6ch
```

