package server

import (
	"fmt"
	"github.com/robertpelloni/tabby/tabby-go/pkg/agent"
	"github.com/robertpelloni/tabby/tabby-go/pkg/vdom"
)

func (s *Server) handleAgentCreateWidget(params interface{}) (*agent.Widget, error) {
	var p struct {
		Type  string `json:"type"`
		Title string `json:"title"`
	}
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return s.agentMgr.CreateWidget(p.Type, p.Title), nil
}

func (s *Server) handleAgentUpdateWidgetVDOM(params interface{}) error {
	var p struct {
		ID   string    `json:"id"`
		VDOM *vdom.Node `json:"vdom"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.agentMgr.UpdateWidgetVDOM(p.ID, p.VDOM)
}
