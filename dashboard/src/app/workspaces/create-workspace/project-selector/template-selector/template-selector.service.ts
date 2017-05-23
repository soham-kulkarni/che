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

import {CheProjectTemplate} from '../../../../../components/api/che-project-template.factory';

/**
 * Service for template selector.
 *
 * @author Oleksii Kurinnyi
 */
export class TemplateSelectorSvc {
  /**
   * Filter service.
   */
  $filter: ng.IFilterService;
  /**
   * Promises service.
   */
  $q: ng.IQService;
  /**
   * Project template API interactions.
   */
  cheProjectTemplate: CheProjectTemplate;
  /**
   * The list of templates.
   */
  templates: Array<che.IProjectTemplate>;
  /**
   * Names of selected templates.
   */
  templateNames: string[];

  /**
   * Default constructor that is using resource injection
   * @ngInject for Dependency injection
   */
  constructor($filter: ng.IFilterService, $q: ng.IQService, cheProjectTemplate: CheProjectTemplate) {
    this.$filter = $filter;
    this.$q = $q;
    this.cheProjectTemplate = cheProjectTemplate;

    this.fetchTemplates();
  }

  /**
   * Fetches list of templates.
   */
  fetchTemplates(): ng.IPromise<any> {
    const defer = this.$q.defer();

    const templates = this.cheProjectTemplate.getAllProjectTemplates();
    if (templates.length) {
      defer.resolve();
    } else {
      this.cheProjectTemplate.fetchTemplates().finally(() => {
        defer.resolve();
      });
    }

    return defer.promise;
  }

  /**
   * Returns list of fetched project templates.
   *
   * @return {Array<che.IProjectTemplate>}
   */
  getTemplates(): Array<che.IProjectTemplate> {
    return this.cheProjectTemplate.getAllProjectTemplates();
  }

  /**
   * Returns project template by name.
   *
   * @param {string} name the project template name
   * @return {undefined|che.IProjectTemplate}
   */
  getTemplateByName(name: string): che.IProjectTemplate {
    return this.getTemplates().find((template: che.IProjectTemplate) => {
      return template.name === name;
    });
  }

  /**
   * Callback which is called when template is checked or unchecked.
   *
   * @param {string[]} templateNames the list of names of selected templates
   */
  onTemplatesSelected(templateNames: string[]): void {
    this.templateNames = templateNames;
  }

  /**
   * Returns selected template name.
   *
   * @return {string[]}
   */
  getTemplateNames(): string[] {
    return this.templateNames;
  }

}
