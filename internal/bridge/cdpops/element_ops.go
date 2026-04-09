package cdpops

import (
	"context"
	"encoding/json"

	"github.com/chromedp/chromedp"
)

func FillByNodeID(ctx context.Context, nodeID int64, value string) error {
	return chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.FromContext(ctx).Target.Execute(ctx, "DOM.focus", map[string]any{"backendNodeId": nodeID}, nil)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var result json.RawMessage
			if err := chromedp.FromContext(ctx).Target.Execute(ctx, "DOM.resolveNode", map[string]any{
				"backendNodeId": nodeID,
			}, &result); err != nil {
				return err
			}
			var resolved struct {
				Object struct {
					ObjectID string `json:"objectId"`
				} `json:"object"`
			}
			if err := json.Unmarshal(result, &resolved); err != nil {
				return err
			}
			// Use the native value setter for the concrete element type to bypass
			// framework-patched setters (e.g. React's value tracker). Calling the
			// input setter on a textarea throws with an incompatible receiver.
			js := `function(v) {
				var proto = null;
				if (this instanceof window.HTMLTextAreaElement) {
					proto = window.HTMLTextAreaElement.prototype;
				} else if (this instanceof window.HTMLInputElement) {
					proto = window.HTMLInputElement.prototype;
				}
				var setter = proto && Object.getOwnPropertyDescriptor(proto, 'value').set;
				if (setter) { setter.call(this, v); } else { this.value = v; }
				this.dispatchEvent(new Event('input', {bubbles: true}));
				this.dispatchEvent(new Event('change', {bubbles: true}));
			}`
			return chromedp.FromContext(ctx).Target.Execute(ctx, "Runtime.callFunctionOn", map[string]any{
				"functionDeclaration": js,
				"objectId":            resolved.Object.ObjectID,
				"arguments":           []map[string]any{{"value": value}},
			}, nil)
		}),
	)
}

func SelectByNodeID(ctx context.Context, nodeID int64, value string) error {
	return chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.FromContext(ctx).Target.Execute(ctx, "DOM.focus", map[string]any{"backendNodeId": nodeID}, nil)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var result json.RawMessage
			if err := chromedp.FromContext(ctx).Target.Execute(ctx, "DOM.resolveNode", map[string]any{
				"backendNodeId": nodeID,
			}, &result); err != nil {
				return err
			}
			var resolved struct {
				Object struct {
					ObjectID string `json:"objectId"`
				} `json:"object"`
			}
			if err := json.Unmarshal(result, &resolved); err != nil {
				return err
			}
			js := `function(v) { this.value = v; this.dispatchEvent(new Event('input', {bubbles: true})); this.dispatchEvent(new Event('change', {bubbles: true})); }`
			return chromedp.FromContext(ctx).Target.Execute(ctx, "Runtime.callFunctionOn", map[string]any{
				"functionDeclaration": js,
				"objectId":            resolved.Object.ObjectID,
				"arguments":           []map[string]any{{"value": value}},
			}, nil)
		}),
	)
}

// ReadInputValue reads back the effective value of an input element. For React
// controlled inputs it checks the fiber's memoizedProps.value (which reflects
// React state) rather than the DOM value, since the DOM value can be stale.
// Returns the effective value the framework considers current.
func ReadInputValue(ctx context.Context, nodeID int64) (string, error) {
	var value string
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var result json.RawMessage
		if err := chromedp.FromContext(ctx).Target.Execute(ctx, "DOM.resolveNode", map[string]any{
			"backendNodeId": nodeID,
		}, &result); err != nil {
			return err
		}
		var resolved struct {
			Object struct {
				ObjectID string `json:"objectId"`
			} `json:"object"`
		}
		if err := json.Unmarshal(result, &resolved); err != nil {
			return err
		}
		js := `function() {
			var el = this;
			var fiberKey = Object.keys(el).find(function(k) {
				return k.startsWith('__reactFiber$') || k.startsWith('__reactInternalInstance$');
			});
			if (fiberKey) {
				var fiber = el[fiberKey];
				var props = fiber && fiber.memoizedProps;
				if (props && 'value' in props) {
					return props.value || "";
				}
			}
			return el.value || "";
		}`
		var callResult json.RawMessage
		if err := chromedp.FromContext(ctx).Target.Execute(ctx, "Runtime.callFunctionOn", map[string]any{
			"functionDeclaration": js,
			"objectId":            resolved.Object.ObjectID,
			"returnByValue":       true,
		}, &callResult); err != nil {
			return err
		}
		var cr struct {
			Result struct {
				Value string `json:"value"`
			} `json:"result"`
		}
		if err := json.Unmarshal(callResult, &cr); err != nil {
			return err
		}
		value = cr.Result.Value
		return nil
	}))
	return value, err
}

func ScrollByNodeID(ctx context.Context, nodeID int64) error {
	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return chromedp.FromContext(ctx).Target.Execute(ctx, "DOM.scrollIntoViewIfNeeded", map[string]any{"backendNodeId": nodeID}, nil)
	}))
}
