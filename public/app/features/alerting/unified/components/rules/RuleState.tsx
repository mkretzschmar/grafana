import { css } from '@emotion/css';
import { GrafanaTheme2, intervalToAbbreviatedDurationString } from '@grafana/data';
import { HorizontalGroup, Spinner, useStyles2 } from '@grafana/ui';
import { CombinedRule } from 'app/types/unified-alerting';
import { PromAlertingRuleState } from 'app/types/unified-alerting-dto';
import React, { FC, useMemo } from 'react';
import { isAlertingRule, isRecordingRule } from '../../utils/rules';
import { AlertStateTag } from './AlertStateTag';

interface Props {
  rule: CombinedRule;
  isDeleting: boolean;
  isCreating: boolean;
}

export const RuleState: FC<Props> = ({ rule, isDeleting, isCreating }) => {
  const style = useStyles2(getStyle);
  const { promRule } = rule;

  // return how long the rule has been in it's firing state, if any
  const forTime = useMemo(() => {
    if (
      promRule &&
      isAlertingRule(promRule) &&
      promRule.alerts?.length &&
      promRule.state !== PromAlertingRuleState.Inactive
    ) {
      // find earliest alert
      const firstActiveAt = promRule.alerts.reduce((prev, alert) => {
        if (alert.activeAt) {
          const activeAt = new Date(alert.activeAt);
          if (prev === null || prev.getTime() > activeAt.getTime()) {
            return activeAt;
          }
        }
        return prev;
      }, null as Date | null);

      // caclulate time elapsed from earliest alert
      if (firstActiveAt) {
        return (
          <span title={String(firstActiveAt)} className={style.for}>
            for{' '}
            {intervalToAbbreviatedDurationString(
              {
                start: firstActiveAt,
                end: new Date(),
              },
              false
            )}
          </span>
        );
      }
    }
    return null;
  }, [promRule, style]);

  if (isDeleting) {
    return (
      <HorizontalGroup>
        <Spinner />
        deleting
      </HorizontalGroup>
    );
  } else if (isCreating) {
    return (
      <HorizontalGroup>
        {' '}
        <Spinner />
        creating
      </HorizontalGroup>
    );
  } else if (promRule && isAlertingRule(promRule)) {
    return (
      <HorizontalGroup>
        <AlertStateTag state={promRule.state} />
        {forTime}
      </HorizontalGroup>
    );
  } else if (promRule && isRecordingRule(promRule)) {
    return <>Recording rule</>;
  }
  return <>n/a</>;
};

const getStyle = (theme: GrafanaTheme2) => ({
  for: css`
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
    white-space: nowrap;
  `,
});
