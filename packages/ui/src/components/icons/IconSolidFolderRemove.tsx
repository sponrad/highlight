import React from 'react'
import { IconProps } from './types'

export const IconSolidFolderRemove = ({
	size = '1em',
	...props
}: IconProps) => {
	return (
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width={size}
			height={size}
			fill="none"
			viewBox="0 0 20 20"
			focusable="false"
			{...props}
		>
			<path
				fill="currentColor"
				fillRule="evenodd"
				d="M4 4a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-5L9 4H4Zm4 6a1 1 0 1 0 0 2h4a1 1 0 1 0 0-2H8Z"
				clipRule="evenodd"
			/>
		</svg>
	)
}
