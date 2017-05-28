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

/**
 * This class is handling the service for the Zip project import.
 *
 * @author Oleksii Kurinnyi
 */
export class ImportZipProjectService {
  /**
   * Zip project location.
   */
  private location: string;

  /**
   * Callback which is called when location is changed.
   *
   * @param location
   */
  onLocationChanged(location: string): void {
    this.location = location;
  }

  /**
   * Returns project's properties.
   *
   * @return {che.IProjectTemplate}
   */
  getProjectProps(): che.IProjectTemplate {
    const props = {} as che.IProjectTemplate;

    const [ , name] = /\/([^\/]+)\.zip$/i.exec(this.location);
    const path = '/' +  name.replace(/[^\w-_]/g, '_');
    props.name = name;
    props.displayName = name;
    props.description = '';
    props.path = path;
    props.category = '';

    props.source = {
      type: 'zip',
      location: this.location
    };

    return props;
  }
}
