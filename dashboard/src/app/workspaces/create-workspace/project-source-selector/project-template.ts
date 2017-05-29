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

export class ProjectTemplate implements che.IProjectTemplate {
  name: string;
  displayName: string;
  description: string;
  source: che.IProjectSource;
  path: string;
  commands: Array<any>;
  projectType: string;
  type: string;
  tags: Array<string>;
  options: Array<any>;
  workspaceId: string;
  workspaceName: string;
  projects: che.IProject[];

  constructor(props: any) {
  }

}
