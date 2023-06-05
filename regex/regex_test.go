package regex

import (
	"io/ioutil"
	"os/exec"
	"reflect"
	"testing"

	"github.com/ylacancellera/galera-log-explainer/types"
	"github.com/ylacancellera/galera-log-explainer/utils"
)

func TestRegexes(t *testing.T) {
	utils.SkipColor = true
	tests := []struct {
		name                 string
		log, expectedOut     string
		inputCtx             types.LogCtx
		expectedCtx          types.LogCtx
		displayerExpectedNil bool
		mapToTest            types.RegexMap
		key                  string
	}{

		{

			log:         "2001-01-01T01:01:01.000000Z 0 [System] [MY-010910] [Server] /usr/sbin/mysqld: Shutdown complete (mysqld 8.0.23-14.1)  Percona XtraDB Cluster (GPL), Release rel14, Revision d3b9a1d, WSREP version 26.4.3.",
			expectedCtx: types.LogCtx{State: "CLOSED"},
			expectedOut: "shutdown complete",
			mapToTest:   EventsMap,
			key:         "RegexShutdownComplete",
		},

		{
			log:         "2001-01-01 01:01:01 140430087788288 [Note] WSREP: /opt/rh-mariadb102/root/usr/libexec/mysqld: Terminated.",
			expectedCtx: types.LogCtx{State: "CLOSED"},
			expectedOut: "terminated",
			mapToTest:   EventsMap,
			key:         "RegexTerminated",
		},
		{
			log:         "2001-01-01T01:01:01.000000Z 8 [Note] WSREP: /usr/sbin/mysqld: Terminated.",
			expectedCtx: types.LogCtx{State: "CLOSED"},
			expectedOut: "terminated",
			mapToTest:   EventsMap,
			key:         "RegexTerminated",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [WSREP] Received shutdown signal. Will sleep for 10 secs before initiating shutdown. pxc_maint_mode switched to SHUTDOWN",
			expectedCtx: types.LogCtx{State: "CLOSED"},
			expectedOut: "received shutdown",
			mapToTest:   EventsMap,
			key:         "RegexShutdownSignal",
		},
		{
			log:         "2001-01-01 01:01:01 139688443508480 [Note] /opt/rh-mariadb102/root/usr/libexec/mysqld (unknown): Normal shutdown",
			expectedCtx: types.LogCtx{State: "CLOSED"},
			expectedOut: "received shutdown",
			mapToTest:   EventsMap,
			key:         "RegexShutdownSignal",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [ERROR] [MY-010119] [Server] Aborting",
			expectedCtx: types.LogCtx{State: "CLOSED"},
			expectedOut: "ABORTING",
			mapToTest:   EventsMap,
			key:         "RegexAborting",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [Galera] wsrep_load(): loading provider library '/usr/lib64/galera4/libgalera_smm.so'",
			expectedCtx: types.LogCtx{State: "OPEN"},
			expectedOut: "started(cluster)",
			mapToTest:   EventsMap,
			key:         "RegexWsrepLoad",
		},
		{
			log:         "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [Galera] wsrep_load(): loading provider library 'none'",
			expectedCtx: types.LogCtx{State: "OPEN"},
			expectedOut: "started(standalone)",
			mapToTest:   EventsMap,
			key:         "RegexWsrepLoad",
		},

		{
			log:         "2001-01-01 01:01:01 140557650536640 [Note] WSREP: wsrep_load(): loading provider library '/opt/rh-mariadb102/root/usr/lib64/galera/libgalera_smm.so'",
			expectedCtx: types.LogCtx{State: "OPEN"},
			expectedOut: "started(cluster)",
			mapToTest:   EventsMap,
			key:         "RegexWsrepLoad",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 3 [Note] [MY-000000] [Galera] Recovered position from storage: 7780bb61-87cf-11eb-b53b-6a7c64b0fee3:23506640",
			expectedCtx: types.LogCtx{State: "RECOVERY"},
			expectedOut: "wsrep recovery",
			mapToTest:   EventsMap,
			key:         "RegexWsrepRecovery",
		},
		{
			log:         " INFO: WSREP: Recovered position 9a4db4a5-5cf1-11ec-940d-6ba8c5905c02:30",
			expectedCtx: types.LogCtx{State: "RECOVERY"},
			expectedOut: "wsrep recovery",
			mapToTest:   EventsMap,
			key:         "RegexWsrepRecovery",
		},
		{
			log:         " INFO: WSREP: Recovered position 00000000-0000-0000-0000-000000000000:-1",
			expectedCtx: types.LogCtx{State: "RECOVERY"},
			expectedOut: "wsrep recovery",
			mapToTest:   EventsMap,
			key:         "RegexWsrepRecovery",
		},

		{
			log:         "2001-01-01T01:01:01.045425-05:00 0 [ERROR] unknown variable 'validate_password_length=8'",
			expectedOut: "unknown variable: validate_password_length=8",
			mapToTest:   EventsMap,
			key:         "RegexUnknownConf",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [ERROR] [MY-013183] [InnoDB] Assertion failure: btr0cur.cc:296:btr_page_get_prev(get_block->frame, mtr) == page_get_page_no(page) thread 139538894652992",
			expectedCtx: types.LogCtx{State: "CLOSED"},
			expectedOut: "ASSERTION FAILURE",
			mapToTest:   EventsMap,
			key:         "RegexAssertionFailure",
		},

		{
			log:         "2001-01-01  5:06:12 47285568576576 [ERROR] WSREP: failed to open gcomm backend connection: 98: error while trying to listen 'tcp://0.0.0.0:4567?socket.non_blocking=1', asio error 'bind: Address already in use': 98 (Address already in use)",
			expectedCtx: types.LogCtx{State: "CLOSED"},
			expectedOut: "bind address already used",
			mapToTest:   EventsMap,
			key:         "RegexBindAddressAlreadyUsed",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [Galera] (90002222-1111, 'ssl://0.0.0.0:4567') Found matching local endpoint for a connection, blacklisting address ssl://127.0.0.1:4567",
			expectedCtx: types.LogCtx{OwnIPs: []string{"127.0.0.1"}},
			expectedOut: "127.0.0.1 is local",
			mapToTest:   IdentsMap,
			key:         "RegexSourceNode",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [Galera] Passing config to GCS: base_dir = /var/lib/mysql/; base_host = 127.0.0.1; base_port = 4567; cert.log_conflicts = no; cert.optimistic_pa = no; debug = no; evs.auto_evict = 0; evs.delay_margin = PT1S; evs.delayed_keep_period = PT30S; evs.inactive_check_period = PT0.5S; evs.inactive_timeout = PT15S; evs.join_retrans_period = PT1S; evs.max_install_timeouts = 3; evs.send_window = 10; evs.stats_report_period = PT1M; evs.suspect_timeout = PT5S; evs.user_send_window = 4; evs.view_forget_timeout = PT24H; gcache.dir = /data/mysql/; gcache.freeze_purge_at_seqno = -1; gcache.keep_pages_count = 0; gcache.keep_pages_size = 0; gcache.mem_size = 0; gcache.name = galera.cache; gcache.page_size = 128M; gcache.recover = yes; gcache.size = 128M; gcomm.thread_prio = ; gcs.fc_debug = 0; gcs.fc_factor = 1.0; gcs.fc_limit = 100; gcs.fc_master_slave = no; gcs.max_packet_size = 64500; gcs.max_throttle = 0.25; gcs.recv_q_hard_limit = 9223372036854775807; gcs.recv_q_soft_limit = 0.25; gcs.sync_donor = no; gmcast.segment = 0; gmcast.version = 0; pc.announce_timeout = PT3S; pc.checksum = false; pc.ignore_quorum = false; pc.ignore_sb = false; pc.npvo = false; pc.recovery = true; pc.version = 0; pc.wait_prim = true; pc.wait_prim_timeout = PT30S; pc.weight = 1; protonet.backend = asio; protonet.version = 0; repl.causal_read_timeout = PT30S; repl.commit_order = 3; repl.key_format = FLAT8; repl.max_ws_size = 2147483647; repl.proto_max = 10; socket.checksum = 2; socket.recv_buf_size = auto; socket.send_buf_size = auto; socket.ssl_ca = ca.pem; socket.ssl_cert = server-cert.pem; socket.ssl_cipher = ; socket.ssl_compression = YES; socket.ssl_key = server-key.pem;",
			expectedCtx: types.LogCtx{OwnIPs: []string{"127.0.0.1"}},
			expectedOut: "127.0.0.1 is local",
			mapToTest:   IdentsMap,
			key:         "RegexBaseHost",
		},

		{
			log: "        0: 015702fc-32f5-11ed-a4ca-267f97316394, node1",
			inputCtx: types.LogCtx{
				MyIdx:          "0",
				State:          "PRIMARY",
				MemberCount:    1,
				OwnHashes:      []string{},
				OwnNames:       []string{},
				HashToNodeName: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				MyIdx:          "0",
				State:          "PRIMARY",
				MemberCount:    1,
				OwnHashes:      []string{"015702fc-a4ca"},
				OwnNames:       []string{"node1"},
				HashToNodeName: map[string]string{"015702fc-a4ca": "node1"},
			},
			expectedOut: "015702fc-a4ca is node1",
			mapToTest:   IdentsMap,
			key:         "RegexMemberAssociations",
		},
		{
			log: "        0: 015702fc-32f5-11ed-a4ca-267f97316394, node1",
			inputCtx: types.LogCtx{
				MyIdx:          "0",
				State:          "NON-PRIMARY",
				MemberCount:    1,
				OwnHashes:      []string{},
				OwnNames:       []string{},
				HashToNodeName: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				MyIdx:          "0",
				State:          "NON-PRIMARY",
				MemberCount:    1,
				OwnHashes:      []string{"015702fc-a4ca"},
				OwnNames:       []string{"node1"},
				HashToNodeName: map[string]string{"015702fc-a4ca": "node1"},
			},
			expectedOut: "015702fc-a4ca is node1",
			mapToTest:   IdentsMap,
			key:         "RegexMemberAssociations",
		},
		{
			log: "        0: 015702fc-32f5-11ed-a4ca-267f97316394, node1",
			inputCtx: types.LogCtx{
				MyIdx:          "0",
				State:          "NON-PRIMARY",
				MemberCount:    2,
				HashToNodeName: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				MyIdx:          "0",
				State:          "NON-PRIMARY",
				MemberCount:    2,
				HashToNodeName: map[string]string{"015702fc-a4ca": "node1"},
			},
			expectedOut: "015702fc-a4ca is node1",
			mapToTest:   IdentsMap,
			key:         "RegexMemberAssociations",
		},
		{
			log: "        1: 015702fc-32f5-11ed-a4ca-267f97316394, node1",
			inputCtx: types.LogCtx{
				MyIdx:          "1",
				State:          "PRIMARY",
				MemberCount:    1,
				OwnHashes:      []string{},
				OwnNames:       []string{},
				HashToNodeName: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				MyIdx:          "1",
				State:          "PRIMARY",
				MemberCount:    1,
				OwnHashes:      []string{"015702fc-a4ca"},
				OwnNames:       []string{"node1"},
				HashToNodeName: map[string]string{"015702fc-a4ca": "node1"},
			},
			expectedOut: "015702fc-a4ca is node1",
			mapToTest:   IdentsMap,
			key:         "RegexMemberAssociations",
		},
		{
			log: "        0: 015702fc-32f5-11ed-a4ca-267f97316394, node1",
			inputCtx: types.LogCtx{
				MyIdx:          "1",
				State:          "PRIMARY",
				MemberCount:    1,
				HashToNodeName: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				MyIdx:          "1",
				State:          "PRIMARY",
				MemberCount:    1,
				HashToNodeName: map[string]string{"015702fc-a4ca": "node1"},
			},
			expectedOut: "015702fc-a4ca is node1",
			mapToTest:   IdentsMap,
			key:         "RegexMemberAssociations",
		},
		{
			log: "        0: 015702fc-32f5-11ed-a4ca-267f97316394, node1.with.complete.fqdn",
			inputCtx: types.LogCtx{
				MyIdx:          "1",
				State:          "PRIMARY",
				MemberCount:    1,
				HashToNodeName: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				MyIdx:          "1",
				State:          "PRIMARY",
				MemberCount:    1,
				HashToNodeName: map[string]string{"015702fc-a4ca": "node1"},
			},
			expectedOut: "015702fc-a4ca is node1",
			mapToTest:   IdentsMap,
			key:         "RegexMemberAssociations",
		},
		{
			name: "name too long and truncated",
			log:  "        0: 015702fc-32f5-11ed-a4ca-267f97316394, name_so_long_it_will_get_trunca",
			inputCtx: types.LogCtx{
				MyIdx:       "1",
				State:       "PRIMARY",
				MemberCount: 1,
			},
			expectedCtx: types.LogCtx{
				MyIdx:       "1",
				State:       "PRIMARY",
				MemberCount: 1,
			},
			expectedOut:          "",
			displayerExpectedNil: true,
			mapToTest:            IdentsMap,
			key:                  "RegexMemberAssociations",
		},

		{
			log:         "  members(1):",
			expectedOut: "view member count: 1",
			expectedCtx: types.LogCtx{MemberCount: 1},
			mapToTest:   IdentsMap,
			key:         "RegexMemberCount",
		},

		{
			log:      "2001-01-01T01:01:01.000000Z 1 [Note] [MY-000000] [Galera] ####### My UUID: 60205de0-5cf6-11ec-8884-3a01908be11a",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				OwnHashes: []string{"60205de0-8884"},
			},
			expectedOut: "60205de0-8884 is local",
			mapToTest:   IdentsMap,
			key:         "RegexOwnUUID",
		},

		{
			log:      "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: (9509c194, 'tcp://0.0.0.0:4567') turning message relay requesting on, nonlive peers:",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				OwnHashes: []string{"9509c194"},
			},
			expectedOut: "9509c194 is local",
			mapToTest:   IdentsMap,
			key:         "RegexOwnUUIDFromMessageRelay",
		},

		{
			log:      "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: New COMPONENT: primary = yes, bootstrap = no, my_idx = 0, memb_num = 2",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				MyIdx: "0",
			},
			expectedOut: "my_idx=0",
			mapToTest:   IdentsMap,
			key:         "RegexMyIDXFromComponent",
		},

		{
			log:      "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: (9509c194, 'tcp://0.0.0.0:4567') connection established to 838ebd6d tcp://172.17.0.2:4567",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				OwnHashes: []string{"9509c194"},
			},
			expectedOut: "9509c194 is local",
			mapToTest:   IdentsMap,
			key:         "RegexOwnUUIDFromEstablished",
		},

		{
			log:      "  own_index: 1",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				MyIdx: "1",
			},
			expectedOut: "my_idx=1",
			mapToTest:   IdentsMap,
			key:         "RegexOwnIndexFromView",
		},

		{
			log: "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [Galera] (60205de0-8884, 'ssl://0.0.0.0:4567') connection established to 5873acd0-baa8 ssl://172.17.0.2:4567",
			inputCtx: types.LogCtx{
				HashToIP: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				HashToIP: map[string]string{"5873acd0-baa8": "172.17.0.2"},
			},
			expectedOut: "172.17.0.2 established",
			mapToTest:   ViewsMap,
			key:         "RegexNodeEstablished",
		},
		{
			name: "established to node's own ip",
			log:  "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [Galera] (60205de0-8884, 'ssl://0.0.0.0:4567') connection established to 5873acd0-baa8 ssl://172.17.0.2:4567",
			inputCtx: types.LogCtx{
				OwnIPs:   []string{"172.17.0.2"},
				HashToIP: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				OwnIPs:   []string{"172.17.0.2"},
				HashToIP: map[string]string{"5873acd0-baa8": "172.17.0.2"},
			},
			expectedOut:          "",
			displayerExpectedNil: true,
			mapToTest:            ViewsMap,
			key:                  "RegexNodeEstablished",
		},

		{
			log: "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [Galera] declaring 5873acd0-baa8 at ssl://172.17.0.2:4567 stable",
			inputCtx: types.LogCtx{
				HashToIP:   map[string]string{},
				IPToMethod: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				HashToIP:   map[string]string{"5873acd0-baa8": "172.17.0.2"},
				IPToMethod: map[string]string{"172.17.0.2": "ssl"},
			},
			expectedOut: "172.17.0.2 joined",
			mapToTest:   ViewsMap,
			key:         "RegexNodeJoined",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [Galera] forgetting 871c35de-99ae (ssl://172.17.0.2:4567)",
			expectedOut: "172.17.0.2 left",
			mapToTest:   ViewsMap,
			key:         "RegexNodeLeft",
		},

		{
			log: "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: New COMPONENT: primary = yes, bootstrap = no, my_idx = 0, memb_num = 2",
			expectedCtx: types.LogCtx{
				State:       "PRIMARY",
				MemberCount: 2,
			},
			expectedOut: "PRIMARY(n=2)",
			mapToTest:   ViewsMap,
			key:         "RegexNewComponent",
		},
		{
			name: "bootstrap",
			log:  "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: New COMPONENT: primary = yes, bootstrap = yes, my_idx = 0, memb_num = 2",
			expectedCtx: types.LogCtx{
				State:       "PRIMARY",
				MemberCount: 2,
			},
			expectedOut: "PRIMARY(n=2),bootstrap",
			mapToTest:   ViewsMap,
			key:         "RegexNewComponent",
		},
		{
			name: "don't set primary",
			log:  "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: New COMPONENT: primary = yes, bootstrap = no, my_idx = 0, memb_num = 2",
			inputCtx: types.LogCtx{
				State: "JOINER",
			},
			expectedCtx: types.LogCtx{
				State:       "JOINER",
				MemberCount: 2,
			},
			expectedOut: "PRIMARY(n=2)",
			mapToTest:   ViewsMap,
			key:         "RegexNewComponent",
		},
		{
			name: "non-primary",
			log:  "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: New COMPONENT: primary = no, bootstrap = no, my_idx = 0, memb_num = 2",
			expectedCtx: types.LogCtx{
				State:       "NON-PRIMARY",
				MemberCount: 2,
			},
			expectedOut: "NON-PRIMARY(n=2)",
			mapToTest:   ViewsMap,
			key:         "RegexNewComponent",
		},

		{
			log: "2001-01-01T01:01:01.000000Z 84580 [Note] [MY-000000] [Galera] evs::proto(9a826787-9e98, LEAVING, view_id(REG,4971d113-87b0,22)) suspecting node: 4971d113-87b0",
			inputCtx: types.LogCtx{
				HashToIP: map[string]string{},
			},
			expectedCtx: types.LogCtx{
				HashToIP: map[string]string{},
			},
			expectedOut: "4971d113-87b0 suspected to be down",
			mapToTest:   ViewsMap,
			key:         "RegexNodeSuspect",
		},
		{
			name: "with known ip",
			log:  "2001-01-01T01:01:01.000000Z 84580 [Note] [MY-000000] [Galera] evs::proto(9a826787-9e98, LEAVING, view_id(REG,4971d113-87b0,22)) suspecting node: 4971d113-87b0",
			inputCtx: types.LogCtx{
				HashToIP: map[string]string{"4971d113-87b0": "172.17.0.2"},
			},
			expectedCtx: types.LogCtx{
				HashToIP: map[string]string{"4971d113-87b0": "172.17.0.2"},
			},
			expectedOut: "172.17.0.2 suspected to be down",
			mapToTest:   ViewsMap,
			key:         "RegexNodeSuspect",
		},

		{
			log: "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: remote endpoint tcp://172.17.0.2:4567 changed identity 84953af9 -> 5a478da2",
			inputCtx: types.LogCtx{
				HashToIP: map[string]string{"84953af9": "172.17.0.2"},
			},
			expectedCtx: types.LogCtx{
				HashToIP: map[string]string{"84953af9": "172.17.0.2", "5a478da2": "172.17.0.2"},
			},
			expectedOut: "172.17.0.2 changed identity",
			mapToTest:   ViewsMap,
			key:         "RegexNodeChangedIdentity",
		},
		{
			name: "with complete uuid",
			log:  "2001-01-01T01:01:01.000000Z 0 [Note] [MY-000000] [Galera] remote endpoint ssl://172.17.0.2:4567 changed identity 595812bc-9c79-11ec-ad3f-3a7953bcc2fc -> 595812bc-9c79-11ec-ad40-3a7953bcc2fc",
			inputCtx: types.LogCtx{
				HashToIP: map[string]string{"595812bc-ad3f": "172.17.0.2"},
			},
			expectedCtx: types.LogCtx{
				HashToIP: map[string]string{"595812bc-ad3f": "172.17.0.2", "595812bc-ad40": "172.17.0.2"},
			},
			expectedOut: "172.17.0.2 changed identity",
			mapToTest:   ViewsMap,
			key:         "RegexNodeChangedIdentity",
		},

		{
			log:      "2001-01-01T01:01:01.000000Z 0 [ERROR] [MY-000000] [Galera] It may not be safe to bootstrap the cluster from this node. It was not the last one to leave the cluster and may not contain all the updates. To force cluster bootstrap with this node, edit the grastate.dat file manually and set safe_to_bootstrap to 1 .",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				State: "CLOSED",
			},
			expectedOut: "not safe to bootstrap",
			mapToTest:   ViewsMap,
			key:         "RegexWsrepUnsafeBootstrap",
		},

		{
			log:      "2001-01-01T01:01:01.481967+09:00 4 [ERROR] WSREP: Node consistency compromised, aborting...",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				State: "CLOSED",
			},
			expectedOut: "consistency compromised",
			mapToTest:   ViewsMap,
			key:         "RegexWsrepConsistenctyCompromised",
		},
		{
			log:      "2001-01-01T01:01:01.000000Z 86 [ERROR] WSREP: Node consistency compromized, aborting...",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				State: "CLOSED",
			},
			expectedOut: "consistency compromised",
			mapToTest:   ViewsMap,
			key:         "RegexWsrepConsistenctyCompromised",
		},

		{
			log:         "2001-01-01  5:06:12 47285568576576 [Note] WSREP: gcomm: bootstrapping new group 'cluster'",
			expectedOut: "bootstrapping",
			mapToTest:   ViewsMap,
			key:         "RegexBootstrap",
		},

		{
			log:      "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: Shifting OPEN -> CLOSED (TO: 1922878)",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				State: "CLOSED",
			},
			expectedOut: "OPEN -> CLOSED",
			mapToTest:   StatesMap,
			key:         "RegexShift",
		},
		{
			log:      "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: Shifting SYNCED -> DONOR/DESYNCED (TO: 21582507)",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				State: "DONOR",
			},
			expectedOut: "SYNCED -> DONOR",
			mapToTest:   StatesMap,
			key:         "RegexShift",
		},
		{
			log:      "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: Shifting DONOR/DESYNCED -> JOINED (TO: 21582507)",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				State: "JOINED",
			},
			expectedOut: "DESYNCED -> JOINED",
			mapToTest:   StatesMap,
			key:         "RegexShift",
		},

		{
			log:      "2001-01-01 01:01:01 140446385440512 [Note] WSREP: Restored state OPEN -> SYNCED (72438094)",
			inputCtx: types.LogCtx{},
			expectedCtx: types.LogCtx{
				State: "SYNCED",
			},
			expectedOut: "(restored)OPEN -> SYNCED",
			mapToTest:   StatesMap,
			key:         "RegexRestoredState",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: Member 2.0 (node2) requested state transfer from '*any*'. Selected 0.0 (node1)(SYNCED) as donor.",
			inputCtx:    types.LogCtx{},
			expectedCtx: types.LogCtx{},
			expectedOut: "node1 will resync node2",
			mapToTest:   SSTMap,
			key:         "RegexSSTRequestSuccess",
		},
		{
			name: "joining",
			log:  "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: Member 2.0 (node2) requested state transfer from '*any*'. Selected 0.0 (node1)(SYNCED) as donor.",
			inputCtx: types.LogCtx{
				OwnNames: []string{"node2"},
			},
			expectedCtx: types.LogCtx{
				OwnNames:         []string{"node2"},
				ResyncedFromNode: "node1",
			},
			expectedOut: "node1 will resync local node",
			mapToTest:   SSTMap,
			key:         "RegexSSTRequestSuccess",
		},
		{
			name: "donor",
			log:  "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: Member 2.0 (node2) requested state transfer from '*any*'. Selected 0.0 (node1)(SYNCED) as donor.",
			inputCtx: types.LogCtx{
				OwnNames: []string{"node1"},
			},
			expectedCtx: types.LogCtx{
				OwnNames:      []string{"node1"},
				ResyncingNode: "node2",
			},
			expectedOut: "local node will resync node2",
			mapToTest:   SSTMap,
			key:         "RegexSSTRequestSuccess",
		},

		{
			log:         "2001-01-01 01:01:01.164  WARN: Member 1.0 (node2) requested state transfer from 'node1', but it is impossible to select State Transfer donor: Resource temporarily unavailable",
			inputCtx:    types.LogCtx{},
			expectedCtx: types.LogCtx{},
			expectedOut: "node2 cannot find donor",
			mapToTest:   SSTMap,
			key:         "RegexSSTResourceUnavailable",
		},
		{
			name: "local",
			log:  "2001-01-01 01:01:01.164  WARN: Member 1.0 (node2) requested state transfer from 'node1', but it is impossible to select State Transfer donor: Resource temporarily unavailable",
			inputCtx: types.LogCtx{
				OwnNames: []string{"node2"},
			},
			expectedCtx: types.LogCtx{
				OwnNames: []string{"node2"},
			},
			expectedOut: "cannot find donor",
			mapToTest:   SSTMap,
			key:         "RegexSSTResourceUnavailable",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: 0.0 (node1): State transfer to 2.0 (node2) complete.",
			inputCtx:    types.LogCtx{},
			expectedCtx: types.LogCtx{},
			expectedOut: "node1 synced node2",
			mapToTest:   SSTMap,
			key:         "RegexSSTComplete",
		},
		{
			name: "joiner",
			log:  "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: 0.0 (node1): State transfer to 2.0 (node2) complete.",
			inputCtx: types.LogCtx{
				OwnNames:         []string{"node2"},
				ResyncedFromNode: "node1",
			},
			expectedCtx: types.LogCtx{
				ResyncedFromNode: "",
				OwnNames:         []string{"node2"},
			},
			expectedOut: "finished resyncing from node1",
			mapToTest:   SSTMap,
			key:         "RegexSSTComplete",
		},
		{
			name: "donor",
			log:  "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: 0.0 (node1): State transfer to 2.0 (node2) complete.",
			inputCtx: types.LogCtx{
				OwnNames:      []string{"node1"},
				ResyncingNode: "node2",
			},
			expectedCtx: types.LogCtx{
				ResyncingNode: "",
				OwnNames:      []string{"node1"},
			},
			expectedOut: "finished sending SST to node2",
			mapToTest:   SSTMap,
			key:         "RegexSSTComplete",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [Note] WSREP: 0.0 (node1): State transfer to -1.-1 (left the group) complete.",
			inputCtx:    types.LogCtx{},
			expectedCtx: types.LogCtx{},
			expectedOut: "node1 synced ??(node left)",
			mapToTest:   SSTMap,
			key:         "RegexSSTCompleteUnknown",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 0 [ERROR] [MY-000000] [WSREP] Process completed with error: wsrep_sst_xtrabackup-v2 --role 'donor' --address '172.17.0.2:4444/xtrabackup_sst//1' --socket '/var/lib/mysql/mysql.sock' --datadir '/var/lib/mysql/' --basedir '/usr/' --plugindir '/usr/lib64/mysql/plugin/' --defaults-file '/etc/my.cnf' --defaults-group-suffix '' --mysqld-version '8.0.28-19.1'   '' --gtid '9db0bcdf-b31a-11ed-a398-2a4cfdd82049:1' : 22 (Invalid argument)",
			expectedOut: "SST error",
			mapToTest:   SSTMap,
			key:         "RegexSSTError",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z 1328586 [Note] [MY-000000] [WSREP] Initiating SST cancellation",
			expectedOut: "Former SST cancelled",
			mapToTest:   SSTMap,
			key:         "RegexSSTCancellation",
		},

		{
			log:         "2001-01-01T01:01:01.000000Z WSREP_SST: [INFO] Proceeding with SST.........",
			expectedCtx: types.LogCtx{State: "JOINER"},
			expectedOut: "Receiving SST",
			mapToTest:   SSTMap,
			key:         "RegexSSTProceeding",
		},

		{
			log: "2001-01-01T01:01:01.000000Z WSREP_SST: [INFO] Streaming the backup to joiner at 172.17.0.2 4444",
			expectedCtx: types.LogCtx{
				State:         "DONOR",
				ResyncingNode: "172.17.0.2",
			},
			expectedOut: "SST to 172.17.0.2",
			mapToTest:   SSTMap,
			key:         "RegexSSTStreamingTo",
		},

		{
			log:         "2001-01-01 01:01:01 140446376740608 [Note] WSREP: IST received: e00c4fff-c4b0-11e9-96a8-0f9789de42ad:69472531",
			expectedCtx: types.LogCtx{},
			expectedOut: "IST received(seqno:69472531)",
			mapToTest:   SSTMap,
			key:         "RegexISTReceived",
		},

		{
			log:         "2001-01-01  7:25:17 140433613571840 [Note] WSREP: async IST sender starting to serve tcp://172.17.0.2:4568 sending 71221242-71221248",
			expectedCtx: types.LogCtx{},
			expectedOut: "IST to 172.17.0.2(seqno:71221248)",
			mapToTest:   SSTMap,
			key:         "RegexISTSender",
		},
	}

	for _, test := range tests {
		if test.name == "" {
			test.name = "default"
		}
		testActualGrepOnLog(t, test.key, test.log, test.mapToTest[test.key])

		ctx, displayer := test.mapToTest[test.key].Handle(test.inputCtx, test.log)
		msg := ""
		if displayer != nil {
			msg = displayer(ctx)
		} else if !test.displayerExpectedNil {
			t.Errorf("key: %s\ntestname: %s\ndisplayer is nil\nexpected: not nil", test.key, test.name)
		}
		if !reflect.DeepEqual(ctx, test.expectedCtx) || msg != test.expectedOut {
			t.Errorf("\nkey: %s\ntestname: %s\nctx: %v\nexpected ctx: %v\nout: %s\nexpected out: %s", test.key, test.name, ctx, test.expectedCtx, msg, test.expectedOut)
			t.Fail()
		}
	}
}

func testActualGrepOnLog(t *testing.T, key, log string, regex *types.LogRegex) {

	f, err := ioutil.TempFile(t.TempDir(), "test_log")
	if err != nil {
		t.Fatalf("failed to create tmp file: %v", err)
	}
	defer f.Sync()

	_, err = f.WriteString(log)
	if err != nil {
		t.Fatalf("failed to write in tmp file: %v", err)
	}
	m := types.RegexMap{"test": regex}

	out, err := exec.Command("grep", "-P", m.Compile()[0], f.Name()).Output()
	if err != nil {
		t.Fatalf("failed to grep in tmp file: %v\nregex key: %s\nregex string: \"%s\"\nlog: %s", err, key, regex.Regex.String(), log)
	}
	if string(out) == "" {
		t.Errorf("empty results when grepping in tmp file: %v, using: %s", err, regex.Regex.String())
	}
}