package channels

import (
	"context"
	"net/url"
	"testing"

	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"
)

func TestLineNotifier(t *testing.T) {
	tmpl := templateForTests(t)

	externalURL, err := url.Parse("http://localhost")
	require.NoError(t, err)
	tmpl.ExternalURL = externalURL

	cases := []struct {
		name         string
		settings     string
		alerts       []*types.Alert
		expHeaders   map[string]string
		expMsg       string
		expInitError error
		expMsgError  error
	}{
		{
			name:     "One alert",
			settings: `{"token": "sometoken"}`,
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1"},
					},
				},
			},
			expHeaders: map[string]string{
				"Authorization": "Bearer sometoken",
				"Content-Type":  "application/x-www-form-urlencoded;charset=UTF-8",
			},
			expMsg:       "message=%5BFIRING%3A1%5D++%28val1%29%0Ahttp%3A%2Flocalhost%2Falerting%2Flist%0A%0A%0A%2A%2AFiring%2A%2A%0ALabels%3A%0A+-+alertname+%3D+alert1%0A+-+lbl1+%3D+val1%0AAnnotations%3A%0A+-+ann1+%3D+annv1%0ASource%3A+%0A%0A%0A%0A%0A",
			expInitError: nil,
			expMsgError:  nil,
		}, {
			name:     "Multiple alerts",
			settings: `{"token": "sometoken"}`,
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1"},
					},
				}, {
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val2"},
						Annotations: model.LabelSet{"ann1": "annv2"},
					},
				},
			},
			expHeaders: map[string]string{
				"Authorization": "Bearer sometoken",
				"Content-Type":  "application/x-www-form-urlencoded;charset=UTF-8",
			},
			expMsg:       "message=%5BFIRING%3A2%5D++%0Ahttp%3A%2Flocalhost%2Falerting%2Flist%0A%0A%0A%2A%2AFiring%2A%2A%0ALabels%3A%0A+-+alertname+%3D+alert1%0A+-+lbl1+%3D+val1%0AAnnotations%3A%0A+-+ann1+%3D+annv1%0ASource%3A+%0ALabels%3A%0A+-+alertname+%3D+alert1%0A+-+lbl1+%3D+val2%0AAnnotations%3A%0A+-+ann1+%3D+annv2%0ASource%3A+%0A%0A%0A%0A%0A",
			expInitError: nil,
			expMsgError:  nil,
		}, {
			name:         "Token missing",
			settings:     `{}`,
			expInitError: alerting.ValidationError{Reason: "Could not find token in settings"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			settingsJSON, err := simplejson.NewJson([]byte(c.settings))
			require.NoError(t, err)

			m := &NotificationChannelConfig{
				Name:     "line_testing",
				Type:     "line",
				Settings: settingsJSON,
			}

			pn, err := NewLineNotifier(m, tmpl)
			if c.expInitError != nil {
				require.Error(t, err)
				require.Equal(t, c.expInitError.Error(), err.Error())
				return
			}
			require.NoError(t, err)

			body := ""
			var headers map[string]string
			bus.AddHandlerCtx("test", func(ctx context.Context, webhook *models.SendWebhookSync) error {
				body = webhook.Body
				headers = webhook.HttpHeader
				return nil
			})

			ctx := notify.WithGroupKey(context.Background(), "alertname")
			ctx = notify.WithGroupLabels(ctx, model.LabelSet{"alertname": ""})
			ok, err := pn.Notify(ctx, c.alerts...)
			if c.expMsgError != nil {
				require.False(t, ok)
				require.Error(t, err)
				require.Equal(t, c.expMsgError.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			require.True(t, ok)

			require.Equal(t, c.expHeaders, headers)
			require.Equal(t, c.expMsg, body)
		})
	}
}
