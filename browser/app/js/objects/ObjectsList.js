/*
 * MinIO Object Storage (c) 2021 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from "react"
import ObjectContainer from "./ObjectContainer"
import PrefixContainer from "./PrefixContainer"

export const ObjectsList = ({ objects }) => {
  const list = objects.map(object => {
    if (object.name.endsWith("/")) {
      return <PrefixContainer object={object} key={object.name} />
    } else {
      return <ObjectContainer object={object} key={object.name} />
    }
  })
  return <div>{list}</div>
}

export default ObjectsList
