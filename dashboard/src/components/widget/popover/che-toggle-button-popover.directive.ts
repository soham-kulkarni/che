/*
 * Copyright (c) 2015-2017 Codenvy, S.A.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *   Codenvy, S.A. - initial API and implementation
 */
'use strict';
import {RandomSvc} from '../../utils/random.service';

interface IPopoverAttrs extends ng.IAttributes {
  buttonTitle: string;
  buttonFontIcon: string;
  buttonOnChange: string;
  buttonState: string;
  chePopoverTitle: string;
  chePopoverPlacement: string;
}

/**
 * @ngdoc directive
 * @name components.directive:CheToggleSinglePopover
 * @restrict E
 * @function
 * @element
 *
 * @description
 * `<toggle-button-popover>` defines two state button's popover.
 *
 * @param {string} button-title button's title
 * @param {string} button-font-icon button's icon CSS class
 * @param {expression=} button-state expression which defines initial state of button.
 * @param {Function} button-on-change callback on model change
 * @param {string=} che-popover-title popover's title
 * @param {string=} che-popover-placement popover's placement
 * @usage
 *   <toggle-button-popover button-title="Filter"
 *                          button-state="ctrl.filterInitState"
 *                          button-on-change="ctrl.filterStateOnChange(state)"><div>My popover</div></toggle-button-popover>
 *
 * @author Oleksii Orel
 */
export class CheToggleButtonPopover implements ng.IDirective {
  restrict = 'E';
  transclude = true;

  private $compile: ng.ICompileService;
  private $timeout: ng.ITimeoutService;
  private randomSvc: RandomSvc;

  /**
   * Default constructor that is using resource
   * @ngInject for Dependency injection
   */
  constructor($compile: ng.ICompileService, $timeout: ng.ITimeoutService, randomSvc: RandomSvc) {
    this.$compile = $compile;
    this.$timeout = $timeout;
    this.randomSvc = randomSvc;
  }

  /**
   * Template for the toggle-button-popover
   * @param {ng.IAugmentedJQuery} element
   * @param {any} attrs
   * @returns {string} the template
   */
  template(element: ng.IAugmentedJQuery, attrs: IPopoverAttrs) {
    return `<toggle-single-button che-title="${attrs.buttonTitle}"
                                  che-font-icon="${attrs.buttonFontIcon}"
                                  che-on-change="${attrs.buttonOnChange}"
                                  che-state="${attrs.buttonState ? attrs.buttonState : 'false'}"
                                  popover-title="${attrs.chePopoverTitle ? attrs.chePopoverTitle : ''}"
                                  popover-placement="${attrs.chePopoverPlacement ? attrs.chePopoverPlacement : 'bottom'}"
                                  popover-is-open="${this.randomSvc.getRandString({prefix: 'isOpenPopover'})}"
                                  uib-popover-html="'<div class=\\'che-transclude\\'></div>'"></toggle-single-button>`;
  }

  link($scope: ng.IScope, $element: ng.IAugmentedJQuery, attrs: IPopoverAttrs, ctrl: any, transclude: ng.ITranscludeFunction): void {
    const jqTranscludeEl = transclude($scope, (clonedElement: ng.IAugmentedJQuery) => {
      return clonedElement;
    });
    if (!jqTranscludeEl.length) {
      return;
    }
    const isOpenAttrKey = $element.find('toggle-single-button').attr('popover-is-open');
    if (!isOpenAttrKey) {
      return;
    }
    const updatePopoverEl = () => {
      if (!$scope[isOpenAttrKey]) {
        return;
      }
      this.$timeout(() => {
        $element.find('.che-transclude').replaceWith(jqTranscludeEl);
        this.$compile(jqTranscludeEl)($scope);
      });
    };
    if (attrs.buttonState === 'true' || $scope[attrs.buttonState]) {
      $scope[isOpenAttrKey] = true;
      updatePopoverEl();
    }
    const watcher = $scope.$watch(($scope: ng.IScope) => {
      return $scope[isOpenAttrKey];
    }, () => {
      updatePopoverEl();
    });
    $scope.$on('$destroy', () => {
      watcher();
    });
  }
}
