// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/johnny-morrice/godless/api (interfaces: RemoteNamespace,RemoteStore,NamespaceTree,NamespaceSearcher,DataPeer,PubSubSubscription,PubSubRecord)

package mock_godless

import (
	gomock "github.com/golang/mock/gomock"
	api "github.com/johnny-morrice/godless/api"
	crdt "github.com/johnny-morrice/godless/crdt"
	query "github.com/johnny-morrice/godless/query"
	io "io"
)

// Mock of RemoteNamespace interface
type MockRemoteNamespace struct {
	ctrl     *gomock.Controller
	recorder *_MockRemoteNamespaceRecorder
}

// Recorder for MockRemoteNamespace (not exported)
type _MockRemoteNamespaceRecorder struct {
	mock *MockRemoteNamespace
}

func NewMockRemoteNamespace(ctrl *gomock.Controller) *MockRemoteNamespace {
	mock := &MockRemoteNamespace{ctrl: ctrl}
	mock.recorder = &_MockRemoteNamespaceRecorder{mock}
	return mock
}

func (_m *MockRemoteNamespace) EXPECT() *_MockRemoteNamespaceRecorder {
	return _m.recorder
}

func (_m *MockRemoteNamespace) Close() {
	_m.ctrl.Call(_m, "Close")
}

func (_mr *_MockRemoteNamespaceRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockRemoteNamespace) Replicate(_param0 []crdt.Link, _param1 api.KvQuery) {
	_m.ctrl.Call(_m, "Replicate", _param0, _param1)
}

func (_mr *_MockRemoteNamespaceRecorder) Replicate(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Replicate", arg0, arg1)
}

func (_m *MockRemoteNamespace) RunKvQuery(_param0 *query.Query, _param1 api.KvQuery) {
	_m.ctrl.Call(_m, "RunKvQuery", _param0, _param1)
}

func (_mr *_MockRemoteNamespaceRecorder) RunKvQuery(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RunKvQuery", arg0, arg1)
}

func (_m *MockRemoteNamespace) RunKvReflection(_param0 api.APIReflectionType, _param1 api.KvQuery) {
	_m.ctrl.Call(_m, "RunKvReflection", _param0, _param1)
}

func (_mr *_MockRemoteNamespaceRecorder) RunKvReflection(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RunKvReflection", arg0, arg1)
}

// Mock of RemoteStore interface
type MockRemoteStore struct {
	ctrl     *gomock.Controller
	recorder *_MockRemoteStoreRecorder
}

// Recorder for MockRemoteStore (not exported)
type _MockRemoteStoreRecorder struct {
	mock *MockRemoteStore
}

func NewMockRemoteStore(ctrl *gomock.Controller) *MockRemoteStore {
	mock := &MockRemoteStore{ctrl: ctrl}
	mock.recorder = &_MockRemoteStoreRecorder{mock}
	return mock
}

func (_m *MockRemoteStore) EXPECT() *_MockRemoteStoreRecorder {
	return _m.recorder
}

func (_m *MockRemoteStore) AddIndex(_param0 crdt.Index) (crdt.IPFSPath, error) {
	ret := _m.ctrl.Call(_m, "AddIndex", _param0)
	ret0, _ := ret[0].(crdt.IPFSPath)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockRemoteStoreRecorder) AddIndex(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AddIndex", arg0)
}

func (_m *MockRemoteStore) AddNamespace(_param0 crdt.Namespace) (crdt.IPFSPath, error) {
	ret := _m.ctrl.Call(_m, "AddNamespace", _param0)
	ret0, _ := ret[0].(crdt.IPFSPath)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockRemoteStoreRecorder) AddNamespace(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AddNamespace", arg0)
}

func (_m *MockRemoteStore) CatIndex(_param0 crdt.IPFSPath) (crdt.Index, error) {
	ret := _m.ctrl.Call(_m, "CatIndex", _param0)
	ret0, _ := ret[0].(crdt.Index)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockRemoteStoreRecorder) CatIndex(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CatIndex", arg0)
}

func (_m *MockRemoteStore) CatNamespace(_param0 crdt.IPFSPath) (crdt.Namespace, error) {
	ret := _m.ctrl.Call(_m, "CatNamespace", _param0)
	ret0, _ := ret[0].(crdt.Namespace)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockRemoteStoreRecorder) CatNamespace(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CatNamespace", arg0)
}

func (_m *MockRemoteStore) Connect() error {
	ret := _m.ctrl.Call(_m, "Connect")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockRemoteStoreRecorder) Connect() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Connect")
}

func (_m *MockRemoteStore) Disconnect() error {
	ret := _m.ctrl.Call(_m, "Disconnect")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockRemoteStoreRecorder) Disconnect() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Disconnect")
}

func (_m *MockRemoteStore) PublishAddr(_param0 crdt.Link, _param1 []api.PubSubTopic) error {
	ret := _m.ctrl.Call(_m, "PublishAddr", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockRemoteStoreRecorder) PublishAddr(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PublishAddr", arg0, arg1)
}

func (_m *MockRemoteStore) SubscribeAddrStream(_param0 api.PubSubTopic) (<-chan crdt.Link, <-chan error) {
	ret := _m.ctrl.Call(_m, "SubscribeAddrStream", _param0)
	ret0, _ := ret[0].(<-chan crdt.Link)
	ret1, _ := ret[1].(<-chan error)
	return ret0, ret1
}

func (_mr *_MockRemoteStoreRecorder) SubscribeAddrStream(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SubscribeAddrStream", arg0)
}

// Mock of NamespaceTree interface
type MockNamespaceTree struct {
	ctrl     *gomock.Controller
	recorder *_MockNamespaceTreeRecorder
}

// Recorder for MockNamespaceTree (not exported)
type _MockNamespaceTreeRecorder struct {
	mock *MockNamespaceTree
}

func NewMockNamespaceTree(ctrl *gomock.Controller) *MockNamespaceTree {
	mock := &MockNamespaceTree{ctrl: ctrl}
	mock.recorder = &_MockNamespaceTreeRecorder{mock}
	return mock
}

func (_m *MockNamespaceTree) EXPECT() *_MockNamespaceTreeRecorder {
	return _m.recorder
}

func (_m *MockNamespaceTree) JoinTable(_param0 crdt.TableName, _param1 crdt.Table) error {
	ret := _m.ctrl.Call(_m, "JoinTable", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockNamespaceTreeRecorder) JoinTable(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "JoinTable", arg0, arg1)
}

func (_m *MockNamespaceTree) LoadTraverse(_param0 api.NamespaceSearcher) error {
	ret := _m.ctrl.Call(_m, "LoadTraverse", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockNamespaceTreeRecorder) LoadTraverse(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "LoadTraverse", arg0)
}

// Mock of NamespaceSearcher interface
type MockNamespaceSearcher struct {
	ctrl     *gomock.Controller
	recorder *_MockNamespaceSearcherRecorder
}

// Recorder for MockNamespaceSearcher (not exported)
type _MockNamespaceSearcherRecorder struct {
	mock *MockNamespaceSearcher
}

func NewMockNamespaceSearcher(ctrl *gomock.Controller) *MockNamespaceSearcher {
	mock := &MockNamespaceSearcher{ctrl: ctrl}
	mock.recorder = &_MockNamespaceSearcherRecorder{mock}
	return mock
}

func (_m *MockNamespaceSearcher) EXPECT() *_MockNamespaceSearcherRecorder {
	return _m.recorder
}

func (_m *MockNamespaceSearcher) ReadNamespace(_param0 crdt.Namespace) api.TraversalUpdate {
	ret := _m.ctrl.Call(_m, "ReadNamespace", _param0)
	ret0, _ := ret[0].(api.TraversalUpdate)
	return ret0
}

func (_mr *_MockNamespaceSearcherRecorder) ReadNamespace(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ReadNamespace", arg0)
}

func (_m *MockNamespaceSearcher) Search(_param0 crdt.Index) []crdt.Link {
	ret := _m.ctrl.Call(_m, "Search", _param0)
	ret0, _ := ret[0].([]crdt.Link)
	return ret0
}

func (_mr *_MockNamespaceSearcherRecorder) Search(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Search", arg0)
}

// Mock of DataPeer interface
type MockDataPeer struct {
	ctrl     *gomock.Controller
	recorder *_MockDataPeerRecorder
}

// Recorder for MockDataPeer (not exported)
type _MockDataPeerRecorder struct {
	mock *MockDataPeer
}

func NewMockDataPeer(ctrl *gomock.Controller) *MockDataPeer {
	mock := &MockDataPeer{ctrl: ctrl}
	mock.recorder = &_MockDataPeerRecorder{mock}
	return mock
}

func (_m *MockDataPeer) EXPECT() *_MockDataPeerRecorder {
	return _m.recorder
}

func (_m *MockDataPeer) Add(_param0 io.Reader) (string, error) {
	ret := _m.ctrl.Call(_m, "Add", _param0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockDataPeerRecorder) Add(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Add", arg0)
}

func (_m *MockDataPeer) Cat(_param0 string) (io.ReadCloser, error) {
	ret := _m.ctrl.Call(_m, "Cat", _param0)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockDataPeerRecorder) Cat(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Cat", arg0)
}

func (_m *MockDataPeer) Connect() error {
	ret := _m.ctrl.Call(_m, "Connect")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockDataPeerRecorder) Connect() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Connect")
}

func (_m *MockDataPeer) Disconnect() error {
	ret := _m.ctrl.Call(_m, "Disconnect")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockDataPeerRecorder) Disconnect() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Disconnect")
}

func (_m *MockDataPeer) IsUp() bool {
	ret := _m.ctrl.Call(_m, "IsUp")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockDataPeerRecorder) IsUp() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsUp")
}

func (_m *MockDataPeer) PubSubPublish(_param0 string, _param1 string) error {
	ret := _m.ctrl.Call(_m, "PubSubPublish", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockDataPeerRecorder) PubSubPublish(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PubSubPublish", arg0, arg1)
}

func (_m *MockDataPeer) PubSubSubscribe(_param0 string) (api.PubSubSubscription, error) {
	ret := _m.ctrl.Call(_m, "PubSubSubscribe", _param0)
	ret0, _ := ret[0].(api.PubSubSubscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockDataPeerRecorder) PubSubSubscribe(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PubSubSubscribe", arg0)
}

// Mock of PubSubSubscription interface
type MockPubSubSubscription struct {
	ctrl     *gomock.Controller
	recorder *_MockPubSubSubscriptionRecorder
}

// Recorder for MockPubSubSubscription (not exported)
type _MockPubSubSubscriptionRecorder struct {
	mock *MockPubSubSubscription
}

func NewMockPubSubSubscription(ctrl *gomock.Controller) *MockPubSubSubscription {
	mock := &MockPubSubSubscription{ctrl: ctrl}
	mock.recorder = &_MockPubSubSubscriptionRecorder{mock}
	return mock
}

func (_m *MockPubSubSubscription) EXPECT() *_MockPubSubSubscriptionRecorder {
	return _m.recorder
}

func (_m *MockPubSubSubscription) Next() (api.PubSubRecord, error) {
	ret := _m.ctrl.Call(_m, "Next")
	ret0, _ := ret[0].(api.PubSubRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockPubSubSubscriptionRecorder) Next() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Next")
}

// Mock of PubSubRecord interface
type MockPubSubRecord struct {
	ctrl     *gomock.Controller
	recorder *_MockPubSubRecordRecorder
}

// Recorder for MockPubSubRecord (not exported)
type _MockPubSubRecordRecorder struct {
	mock *MockPubSubRecord
}

func NewMockPubSubRecord(ctrl *gomock.Controller) *MockPubSubRecord {
	mock := &MockPubSubRecord{ctrl: ctrl}
	mock.recorder = &_MockPubSubRecordRecorder{mock}
	return mock
}

func (_m *MockPubSubRecord) EXPECT() *_MockPubSubRecordRecorder {
	return _m.recorder
}

func (_m *MockPubSubRecord) Data() []byte {
	ret := _m.ctrl.Call(_m, "Data")
	ret0, _ := ret[0].([]byte)
	return ret0
}

func (_mr *_MockPubSubRecordRecorder) Data() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Data")
}

func (_m *MockPubSubRecord) From() string {
	ret := _m.ctrl.Call(_m, "From")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockPubSubRecordRecorder) From() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "From")
}

func (_m *MockPubSubRecord) SeqNo() int64 {
	ret := _m.ctrl.Call(_m, "SeqNo")
	ret0, _ := ret[0].(int64)
	return ret0
}

func (_mr *_MockPubSubRecordRecorder) SeqNo() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SeqNo")
}

func (_m *MockPubSubRecord) TopicIDs() []string {
	ret := _m.ctrl.Call(_m, "TopicIDs")
	ret0, _ := ret[0].([]string)
	return ret0
}

func (_mr *_MockPubSubRecordRecorder) TopicIDs() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "TopicIDs")
}
