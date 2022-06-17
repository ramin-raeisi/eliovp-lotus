module github.com/filecoin-project/lotus

go 1.16

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.1
	contrib.go.opencensus.io/exporter/prometheus v0.4.0
	github.com/BurntSushi/toml v0.4.1
	github.com/GeertJohan/go.rice v1.0.2
	github.com/Gurpartap/async v0.0.0-20180927173644-4f7f499dd9ee
	github.com/Kubuxu/imtui v0.0.0-20210401140320-41663d68d0fa
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/alecthomas/jsonschema v0.0.0-20200530073317-71f438968921
	github.com/buger/goterm v1.0.3
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/containerd/cgroups v0.0.0-20201119153540-4cbc285b3327
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/detailyang/go-fallocate v0.0.0-20180908115635-432fa640bd2e
	github.com/dgraph-io/badger/v2 v2.2007.3
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/docker/go-units v0.4.0
	github.com/drand/drand v1.3.0
	github.com/drand/kyber v1.1.7
	github.com/dustin/go-humanize v1.0.0
	github.com/elastic/go-sysinfo v1.7.0
	github.com/elastic/gosigar v0.14.1
	github.com/etclabscore/go-openrpc-reflect v0.0.36
	github.com/fatih/color v1.13.0
	github.com/filecoin-project/dagstore v0.4.4
	github.com/filecoin-project/filecoin-ffi v0.30.4-0.20200910194244-f640612a1a1f
	github.com/filecoin-project/go-address v0.0.6
	github.com/filecoin-project/go-bitfield v0.2.4
	github.com/filecoin-project/go-cbor-util v0.0.1
	github.com/filecoin-project/go-commp-utils v0.1.3
	github.com/filecoin-project/go-crypto v0.0.1
	github.com/filecoin-project/go-data-transfer v1.12.0
	github.com/filecoin-project/go-fil-commcid v0.1.0
	github.com/filecoin-project/go-fil-commp-hashhash v0.1.0
	github.com/filecoin-project/go-fil-markets v1.13.5
	github.com/filecoin-project/go-jsonrpc v0.1.5
	github.com/filecoin-project/go-padreader v0.0.1
	github.com/filecoin-project/go-paramfetch v0.0.3-0.20220111000201-e42866db1a53
	github.com/filecoin-project/go-state-types v0.1.3
	github.com/filecoin-project/go-statemachine v1.0.1
	github.com/filecoin-project/go-statestore v0.2.0
	github.com/filecoin-project/go-storedcounter v0.1.0
	github.com/filecoin-project/specs-actors v0.9.14
	github.com/filecoin-project/specs-actors/v2 v2.3.6
	github.com/filecoin-project/specs-actors/v3 v3.1.1
	github.com/filecoin-project/specs-actors/v4 v4.0.1
	github.com/filecoin-project/specs-actors/v5 v5.0.4
	github.com/filecoin-project/specs-actors/v6 v6.0.1
	github.com/filecoin-project/specs-actors/v7 v7.0.0-rc1
	github.com/filecoin-project/specs-storage v0.1.1-0.20211228030229-6d460d25a0c9
	github.com/filecoin-project/test-vectors/schema v0.0.5
	github.com/gbrlsnchs/jwt/v3 v3.0.1
	github.com/gdamore/tcell/v2 v2.2.0
	github.com/go-kit/kit v0.12.0
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/hako/durafmt v0.0.0-20200710122514-c0fb7b4da026
	github.com/hannahhoward/go-pubsub v0.0.0-20200423002714-8d62886cc36e
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/icza/backscanner v0.0.0-20210726202459-ac2ffc679f94
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab
	github.com/ipfs/bbloom v0.0.4
	github.com/ipfs/go-bitswap v0.5.1
	github.com/ipfs/go-block-format v0.0.3
	github.com/ipfs/go-blockservice v0.2.1
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/go-cidutil v0.0.2
	github.com/ipfs/go-datastore v0.5.1
	github.com/ipfs/go-ds-badger2 v0.1.2
	github.com/ipfs/go-ds-leveldb v0.5.0
	github.com/ipfs/go-ds-measure v0.2.0
	github.com/ipfs/go-fs-lock v0.0.6
	github.com/ipfs/go-graphsync v0.11.0
	github.com/ipfs/go-ipfs-blockstore v1.1.2
	github.com/ipfs/go-ipfs-blocksutil v0.0.1
	github.com/ipfs/go-ipfs-chunker v0.0.5
	github.com/ipfs/go-ipfs-ds-help v1.1.0
	github.com/ipfs/go-ipfs-exchange-interface v0.1.0
	github.com/ipfs/go-ipfs-exchange-offline v0.1.1
	github.com/ipfs/go-ipfs-files v0.0.9
	github.com/ipfs/go-ipfs-http-client v0.0.6
	github.com/ipfs/go-ipfs-routing v0.2.1
	github.com/ipfs/go-ipfs-util v0.0.2
	github.com/ipfs/go-ipld-cbor v0.0.6
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-ipld-legacy v0.1.1 // indirect
	github.com/ipfs/go-log/v2 v2.4.0
	github.com/ipfs/go-merkledag v0.5.1
	github.com/ipfs/go-metrics-interface v0.0.1
	github.com/ipfs/go-metrics-prometheus v0.0.2
	github.com/ipfs/go-path v0.0.7
	github.com/ipfs/go-unixfs v0.2.6
	github.com/ipfs/interface-go-ipfs-core v0.4.0
	github.com/ipld/go-car v0.3.3
	github.com/ipld/go-car/v2 v2.1.1
	github.com/ipld/go-codec-dagpb v1.3.0
	github.com/ipld/go-ipld-prime v0.14.3
	github.com/ipld/go-ipld-selector-text-lite v0.0.1
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/libp2p/go-buffer-pool v0.0.2
	github.com/libp2p/go-eventbus v0.2.1
	github.com/libp2p/go-libp2p v0.17.0
	github.com/libp2p/go-libp2p-connmgr v0.3.0
	github.com/libp2p/go-libp2p-core v0.13.0
	github.com/libp2p/go-libp2p-discovery v0.6.0
	github.com/libp2p/go-libp2p-kad-dht v0.15.0
	github.com/libp2p/go-libp2p-noise v0.3.0
	github.com/libp2p/go-libp2p-peerstore v0.6.0
	github.com/libp2p/go-libp2p-pubsub v0.6.0
	github.com/libp2p/go-libp2p-quic-transport v0.15.2
	github.com/libp2p/go-libp2p-record v0.1.3
	github.com/libp2p/go-libp2p-routing-helpers v0.2.3
	github.com/libp2p/go-libp2p-swarm v0.9.0
	github.com/libp2p/go-libp2p-tls v0.3.1
	github.com/libp2p/go-libp2p-yamux v0.7.0
	github.com/libp2p/go-maddr-filter v0.1.0
	github.com/mattn/go-isatty v0.0.14
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-base32 v0.0.4
	github.com/multiformats/go-multiaddr v0.4.1
	github.com/multiformats/go-multiaddr-dns v0.3.1
	github.com/multiformats/go-multibase v0.0.3
	github.com/multiformats/go-multihash v0.1.0
	github.com/multiformats/go-varint v0.0.6
	github.com/open-rpc/meta-schema v0.0.0-20201029221707-1b72ef2ea333
	github.com/opentracing/opentracing-go v1.2.0
	github.com/otiai10/copy v1.5.0
	github.com/polydawn/refmt v0.0.0-20201211092308-30ac6d18308e
	github.com/prometheus/client_golang v1.11.0
	github.com/raulk/clock v1.1.0
	github.com/raulk/go-watchdog v1.2.0
	github.com/streadway/quantile v0.0.0-20150917103942-b0c588724d25
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/urfave/cli/v2 v2.2.0
	github.com/whyrusleeping/bencher v0.0.0-20190829221104-bb6607aa8bba
	github.com/whyrusleeping/cbor-gen v0.0.0-20210713220151-be142a5ae1a8
	github.com/whyrusleeping/ledger-filecoin-go v0.9.1-0.20201010031517-c3dcc1bddce4
	github.com/whyrusleeping/multiaddr-filter v0.0.0-20160516205228-e903e4adabd7
	github.com/whyrusleeping/pubsub v0.0.0-20190708150250-92bcb0691325
	github.com/xorcare/golden v0.6.1-0.20191112154924-b87f686d7542
	go.opencensus.io v0.23.0
	go.uber.org/dig v1.10.0 // indirect
	go.uber.org/fx v1.9.0
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20210917221730-978cfadd31cf
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	golang.org/x/tools v0.1.7
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/cheggaaa/pb.v1 v1.0.28
	gotest.tools v2.2.0+incompatible
	lukechampine.com/blake3 v1.1.7 // indirect
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi

replace github.com/filecoin-project/test-vectors => ./extern/test-vectors

//replace github.com/filecoin-project/specs-actors/v7 => /Users/zenground0/pl/repos/specs-actors

// replace github.com/filecon-project/specs-storage => /Users/zenground0/pl/repos/specs-storage