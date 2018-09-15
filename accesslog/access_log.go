package main

import (
	data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
	als "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
	"github.com/sirupsen/logrus"
	"github.com/tommy351/envoy-control-plane-examples/util"
)

type AccessLogServer struct{}

func (a *AccessLogServer) StreamAccessLogs(stream als.AccessLogService_StreamAccessLogsServer) error {
	for {
		msg, err := stream.Recv()

		if err != nil {
			return err
		}

		switch entries := msg.LogEntries.(type) {
		case *als.StreamAccessLogsMessage_HttpLogs:
			a.printHTTPLogs(entries.HttpLogs)
		case *als.StreamAccessLogsMessage_TcpLogs:
			a.printTCPLogs(entries.TcpLogs)
		}
	}

	return nil
}

func (a *AccessLogServer) printHTTPLogs(entries *als.StreamAccessLogsMessage_HTTPAccessLogEntries) {
	for _, entry := range entries.LogEntry {
		util.Logger.
			WithFields(a.commonFields(entry.CommonProperties)).
			WithFields(a.requestFields(entry.Request)).
			WithFields(a.responseFields(entry.Response)).
			Infoln("HTTP")
	}
}

func (a *AccessLogServer) printTCPLogs(entries *als.StreamAccessLogsMessage_TCPAccessLogEntries) {
	for _, entry := range entries.LogEntry {
		util.Logger.WithFields(a.commonFields(entry.CommonProperties)).Infoln("TCP")
	}
}

func (*AccessLogServer) commonFields(common *data.AccessLogCommon) logrus.Fields {
	return logrus.Fields{
		"start":    common.StartTime,
		"upstream": common.UpstreamCluster,
	}
}

func (*AccessLogServer) requestFields(req *data.HTTPRequestProperties) logrus.Fields {
	return logrus.Fields{
		"method": req.RequestMethod.String(),
		"path":   req.Path,
		"id":     req.RequestId,
	}
}

func (*AccessLogServer) responseFields(res *data.HTTPResponseProperties) logrus.Fields {
	fields := logrus.Fields{}

	if code := res.ResponseCode; code != nil {
		fields["status"] = code.Value
	}

	return fields
}
