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

import {TemplateSelectorSvc} from './template-selector.service';
import {StackSelectorSvc} from '../../stack-selector/stack-selector.service';

export class TemplateSelectorController {
  /**
   * Filter service.
   */
  $filter: ng.IFilterService;
  /**
   * Template selector service.
   */
  templateSelectorSvc: TemplateSelectorSvc;
  /**
   * Stack selector service.
   */
  stackSelectorSvc: StackSelectorSvc;
  /**
   * Helper for lists.
   */
  cheListHelper: che.widget.ICheListHelper;
  /**
   * The list of tags of selected stack.
   */
  stackTags: string[];
  /**
   * Sorted list of templates.
   */
  templates: Array<che.IProjectTemplate>;
  /**
   * Filtered and sorted list of templates.
   */
  filteredTemplates: Array<che.IProjectTemplate>;

  /**
   * Default constructor that is using resource injection
   * @ngInject for Dependency injection
   */
  constructor($filter: ng.IFilterService, $scope: ng.IScope, templateSelectorSvc: TemplateSelectorSvc, stackSelectorSvc: StackSelectorSvc, cheListHelperFactory: che.widget.ICheListHelperFactory) {
    this.$filter = $filter;
    this.templateSelectorSvc = templateSelectorSvc;
    this.stackSelectorSvc = stackSelectorSvc;

    const helperId = 'template-selector';
    this.cheListHelper = cheListHelperFactory.getHelper(helperId);
    $scope.$on('$destroy', () => {
      cheListHelperFactory.removeHelper(helperId);
    });

    this.templates = [];
    this.filteredTemplates = [];

    this.onStackChanged();
    this.stackSelectorSvc.subscribe(this.onStackChanged.bind(this));

    this.templateSelectorSvc.fetchTemplates().then(() => {
      this.templates = this.$filter('orderBy')(this.templateSelectorSvc.getAllTemplates(), ['projectType', 'displayName']);
      this.filterAndSortTemplates();
    });
  }

  /**
   * Callback which is called when stack is selected.
   */
  onStackChanged(): void {
    const stackId = this.stackSelectorSvc.getStackId();
    if (!stackId) {
      return;
    }

    const stack = this.stackSelectorSvc.getStackById(stackId);
    this.stackTags = stack ? stack.tags : [];

    this.filterAndSortTemplates();
  }

  /**
   * Filters templates by tags and sort them by project type and template name.
   */
  filterAndSortTemplates(): void {
    const stackTags = !this.stackTags ? [] : this.stackTags.map((tag: string) => tag.toLowerCase());

    if (stackTags.length) {
      this.filteredTemplates = this.templates.filter((template: che.IProjectTemplate) => {
        const templateTags = template.tags.map((tag: string) => tag.toLowerCase());
        return stackTags.some((tag: string) => templateTags.indexOf(tag) > -1);
      });
    }

    this.cheListHelper.setList(this.filteredTemplates, 'name');
  }

  /**
   * Callback which is when the template checkbox is clicked.
   *
   * @param {string} templateName the project template's name
   * @param {boolean} isChecked <code>true</code> if template's checkbox is checked.
   */
  onTemplateClicked(templateName: string, isChecked: boolean): void {
    this.cheListHelper.itemsSelectionStatus[templateName] = isChecked;

    const selectedTemplates = this.cheListHelper.getSelectedItems() as Array<che.IProjectTemplate>;

    this.templateSelectorSvc.onTemplatesSelected(selectedTemplates);
  }
}
