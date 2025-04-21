package core

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type clusterManager struct {
	client     *clientv3.Client
	session    *concurrency.Session
	election   *concurrency.Election
	nodeID     string
	masterChan chan string
	isMaster   bool
}

const (
	electionKey   = "/dissect/master-election"
	heartbeatKey  = "/dissect/nodes/"
	leaseTTL      = 5
	heartbeatFreq = 2 * time.Second
)

func NewClusterManager(endpoints []string, nodeID string) (*clusterManager, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	session, err := concurrency.NewSession(client, concurrency.WithTTL(leaseTTL))
	if err != nil {
		return nil, err
	}

	return &clusterManager{
		client:     client,
		session:    session,
		election:   concurrency.NewElection(session, electionKey),
		nodeID:     nodeID,
		masterChan: make(chan string, 1),
	}, nil
}

// 参选
func (cm *clusterManager) Campaign(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := cm.election.Campaign(ctx, cm.nodeID); err == nil {
					cm.isMaster = true
					go cm.maintainLeadership(ctx)
					return
				} else {
					logrus.Error("Election participation failed:", err)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	time.Sleep(heartbeatFreq) // 选举过程非master会进行阻塞 手动延迟选举过程
}

// 主节点保活
func (cm *clusterManager) maintainLeadership(ctx context.Context) {
	ticker := time.NewTicker(heartbeatFreq)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := cm.client.Put(ctx,
				heartbeatKey+cm.nodeID,
				time.Now().Format(time.RFC3339),
				clientv3.WithLease(cm.session.Lease())); err != nil {
				cm.isMaster = false
				logrus.Errorln("the lease of the current master node is incorrect.", err)
				return
			}
		}
	}
}

// 监听主节点
func (cm *clusterManager) WatchMaster(ctx context.Context) {
	ch := cm.election.Observe(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case resp := <-ch:
				if len(resp.Kvs) > 0 {
					cm.masterChan <- string(resp.Kvs[0].Value)
				}
			}
		}
	}()
}

// 当前是否主节点
func (cm *clusterManager) IsMaster() bool {
	return cm.isMaster
}

// 获取当前主节点地址
func (cm *clusterManager) GetMaster() (string, error) {
	resp, err := cm.election.Leader(context.Background())
	if err != nil {
		return "", err
	}
	return string(resp.Kvs[0].Value), nil
}

// 获取全部节点
func (cm *clusterManager) GetAll() ([]*mvccpb.KeyValue, error) {
	resp, err := cm.election.Leader(context.Background())
	if err != nil {
		return nil, err
	}
	return resp.Kvs, nil
}
